// Copyright (c) 2020. Temple3x (temple3x@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Aliaksandr Valialkin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This file contains code derived from gorpc.
// The main logic & codes are copied from gorpc.

package xtcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/xerrors"

	"github.com/templexxx/xhex"
	"github.com/zaibyte/pkg/xstrconv"

	"github.com/zaibyte/pkg/uid"

	"github.com/zaibyte/pkg/xrpc"

	"github.com/zaibyte/pkg/xdigest"

	"github.com/zaibyte/pkg/xlog"
)

// Client implements xtcp client.
//
// The client must be started with Client.Start() before use.
//
// It is absolutely safe and encouraged using a single client across arbitrary
// number of concurrently running goroutines.
//
// Default client settings are optimized for high load, so don't override
// them without valid reason.
type Client struct {
	// Server address to connect to.
	Addr string

	// The number of concurrent connections the client should establish
	// to the sever.
	// Default is DefaultClientConns.
	Conns int

	// The maximum number of pending requests in the queue.
	//
	// The number of pending requests should exceed the expected number
	// of concurrent goroutines calling client's methods.
	// Otherwise a lot of xrpc.ErrRequestQueueOverflow errors may appear.
	//
	// Default is DefaultPendingMessages.
	PendingRequests int

	// Maximum request time.
	// Default value is DefaultRequestTimeout.
	RequestTimeout time.Duration

	// Size of send buffer per each underlying connection in bytes.
	// Default value is DefaultClientSendBufferSize.
	SendBufferSize int

	// Size of recv buffer per each underlying connection in bytes.
	// Default value is DefaultClientRecvBufferSize.
	RecvBufferSize int

	// Delay between request flushes.
	//
	// Negative values lead to immediate requests' sending to the server
	// without their buffering. This minimizes rpc latency at the cost
	// of higher CPU and network usage.
	//
	// Default value is DefaultFlushDelay.
	FlushDelay time.Duration

	// The client calls this callback when it needs new connection
	// to the server.
	// The client passes Client.Addr into Dial().
	//
	// By default it returns TCP connections established to the Client.Addr.
	Dial DialFunc

	// If TLS is enabled, encrypted will be set true, otherwise it's false.
	// When it's false, xtcp will check the request/response body by the checksum.
	encrypted bool

	requestsChan chan *asyncResult

	stopChan chan struct{}
	stopWg   sync.WaitGroup
}

// asyncResult is a result returned from Client.callAsync().
type asyncResult struct {
	method  uint8
	reqid   uint64
	oid     [16]byte
	reqData xrpc.Byteser // It only can be released after writing or writing failed.

	respBody xrpc.Byteser
	// resp is become available after <-done unblocks.
	done chan struct{}
	// The error can be read only after <-Done unblocks.
	err error

	canceled uint32
}

// cancel cancels async call.
//
// Canceled call isn't sent to the server unless it is already sent there.
// Canceled call may successfully complete if it has been already sent
// to the server before cancel call.
//
// It is safe calling this function multiple times from concurrently
// running goroutines.
func (r *asyncResult) cancel() {
	atomic.StoreUint32(&r.canceled, 1)
}

func (r *asyncResult) isCanceled() bool {
	return atomic.LoadUint32(&r.canceled) != 0
}

const (
	// DefaultRequestTimeout is the default timeout for client request.
	DefaultRequestTimeout = 5 * time.Second

	// DefaultClientSendBufferSize is the default size for Client send buffers.
	DefaultClientSendBufferSize = 64 * 1024

	// DefaultClientRecvBufferSize is the default size for Client receive buffers.
	DefaultClientRecvBufferSize = 64 * 1024

	// DefaultClientConns is the default connection numbers for Client.
	DefaultClientConns = 4
)

// Start starts rpc client. Establishes connection to the server on Client.Addr.
func (c *Client) Start() error {

	if c.stopChan != nil {
		xlog.Panic("already started")
	}

	if c.PendingRequests <= 0 {
		c.PendingRequests = DefaultPendingMessages
	}
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = DefaultRequestTimeout
	}
	if c.SendBufferSize <= 0 {
		c.SendBufferSize = DefaultClientSendBufferSize
	}
	if c.RecvBufferSize <= 0 {
		c.RecvBufferSize = DefaultClientRecvBufferSize
	}
	if c.FlushDelay == 0 {
		c.FlushDelay = DefaultFlushDelay
	}

	c.requestsChan = make(chan *asyncResult, c.PendingRequests)
	c.stopChan = make(chan struct{})

	if c.Conns <= 0 {
		c.Conns = DefaultClientConns
	}
	if c.Dial == nil {
		c.Dial = defaultDial
	}

	for i := 0; i < c.Conns; i++ {
		c.stopWg.Add(1)
		go clientHandler(c)
	}
	return nil
}

// Stop stops rpc client. Stopped client can be started again.
func (c *Client) Stop() error {
	if c.stopChan == nil {
		xlog.Panic("client must be started before stopping it")
	}
	close(c.stopChan)
	c.stopWg.Wait()
	c.stopChan = nil
	return nil
}

// Put puts object to the ZBuf node which Objecter connected.
func (c *Client) PutObj(reqid uint64, oid string, objData []byte, timeout time.Duration) error {
	_, err := c.callTimeout(reqid, objPutMethod, oid, objData, timeout)
	return err
}

// Get gets object from the ZBuf node which Objecter connected.
func (c *Client) GetObj(reqid uint64, oid string, timeout time.Duration) (obj io.ReadCloser, err error) {
	return c.callTimeout(reqid, objGetMethod, oid, nil, timeout)
}

// Delete deletes object in the ZBuf node which Objecter connected.
func (c *Client) DeleteObj(reqid uint64, oid string, timeout time.Duration) error {
	_, err := c.callTimeout(reqid, objDelMethod, oid, nil, timeout)
	return err
}

// callTimeout sends the given request to the server and obtains response
// from the server.
//
// Returns non-nil error if the response cannot be obtained.
//
// Don't forget starting the client with Client.Start() before calling Client.call().
func (c *Client) callTimeout(reqid uint64, method uint8, oid string, body []byte, timeout time.Duration) (resp xrpc.Byteser, err error) {

	if timeout == 0 {
		timeout = c.RequestTimeout
	}

	var ob [16]byte // Using byte array to save function stack space.
	err = xhex.Decode(ob[:16], xstrconv.ToBytes(oid))
	if err != nil {
		return
	}

	var ar *asyncResult
	if ar, err = c.callAsync(reqid, method, ob, body); err != nil {
		return nil, err
	}

	t := acquireTimer(timeout)

	select {
	case <-ar.done:
		resp, err = ar.respBody, ar.err
		if err != nil {
			if resp != nil {
				_ = resp.Close() // In case of passing resp when there is an error.
			}
			resp = nil
		}

		releaseAsyncResult(ar)
	case <-t.C:
		// Cancel will be captured in write preparation, asyncResult will be released there.
		// Or it has been sent, just waiting for the response.
		//
		// If write broken, ar may not be put back to the pool.
		ar.cancel()
		resp = nil
		err = xrpc.ErrTimeout
	}

	releaseTimer(t)
	return
}

func (c *Client) callAsync(reqid uint64, method uint8, oid [16]byte, body []byte) (ar *asyncResult, err error) {

	if reqid == 0 {
		reqid = uid.MakeReqID()
	}

	if method == 0 || method > 3 {
		return nil, xrpc.ErrNotImplemented
	}

	ar = acquireAsyncResult()

	// In case of after returning, the data will still in the client write process,
	// it may sent dirty data. Copy is a safe choice.
	// TODO better way to avoiding copy.
	if body != nil {
		reqData := xrpc.GetNBytes(len(body))
		_, _ = reqData.Write(body)
		ar.reqData = reqData
	}

	ar.reqid = reqid
	ar.method = method
	ar.oid = oid
	ar.done = make(chan struct{})

	select {
	case c.requestsChan <- ar:
		return ar, nil
	default:
		// Try substituting the oldest async request by the new one
		// on requests' queue overflow.
		// This increases the chances for new request to succeed
		// without timeout.
		select {
		case ar2 := <-c.requestsChan:
			if ar2.done != nil {
				ar2.err = xrpc.ErrRequestQueueOverflow
				close(ar2.done)
			} else {
				releaseAsyncResult(ar2)
			}
		default:
		}

		// After pop, try to put again.
		select {
		case c.requestsChan <- ar:
			return ar, nil
		default:
			// RequestsChan is filled, release it since m wasn't exposed to the caller yet.
			releaseAsyncResult(ar)
			return nil, xrpc.ErrRequestQueueOverflow
		}
	}
}

func clientHandler(c *Client) {
	defer c.stopWg.Done()

	var conn net.Conn
	var err error
	var stopping atomic.Value

	for {
		dialChan := make(chan struct{})
		go func() {
			if conn, err = c.Dial(c.Addr); err != nil {
				if stopping.Load() == nil {
					xlog.Errorf("cannot establish rpc connection to: %s: %s", c.Addr, err)
				}
			}
			close(dialChan)
		}()

		select {
		case <-c.stopChan:
			stopping.Store(true)
			<-dialChan
			return
		case <-dialChan:
		}

		if err != nil {
			select {
			case <-c.stopChan:
				return
			case <-time.After(300 * time.Millisecond): // After 300ms, try to dial again.
			}
			continue
		}

		clientHandleConnection(c, conn)

		select {
		case <-c.stopChan:
			return
		default:
		}
	}
}

func clientHandleConnection(c *Client, conn net.Conn) {

	err := sendHandshake(conn)
	if err != nil {
		xlog.Errorf("failed to handshake to: %s: %s", c.Addr, err)
		_ = conn.Close()
		return
	}

	stopChan := make(chan struct{})

	pendingRequests := make(map[uint64]*asyncResult)
	var pendingRequestsLock sync.Mutex // Only two goroutine here, map with mutex is faster than sync.Map.

	writerDone := make(chan error, 1)
	go clientWriter(c, conn, pendingRequests, &pendingRequestsLock, stopChan, writerDone)

	readerDone := make(chan error, 1)
	go clientReader(c, conn, pendingRequests, &pendingRequestsLock, readerDone)

	select {
	case err = <-writerDone:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
	case err = <-readerDone:
		close(stopChan)
		_ = conn.Close()
		<-writerDone
	case <-c.stopChan:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
		<-writerDone
	}

	for _, ar := range pendingRequests {
		ar.err = err
		if ar.done != nil {
			close(ar.done)
		}
	}
}

func sendHandshake(conn net.Conn) error {
	err := conn.SetWriteDeadline(time.Now().Add(handshakeDuration))
	if err != nil {
		return err
	}
	_, err = conn.Write(handshake[:])
	if err != nil {
		return err
	}
	return nil
}

func clientWriter(c *Client, w net.Conn, pendingRequests map[uint64]*asyncResult, pendingRequestsLock *sync.Mutex,
	stopChan <-chan struct{}, done chan<- error) {

	var err error
	defer func() { done <- err }()

	enc := newEncoder(w, c.SendBufferSize)
	msg := new(message)
	header := new(reqHeader)

	t := time.NewTimer(c.FlushDelay)
	var flushChan <-chan time.Time

	var msgID uint64 = 1

	for {
		msgID++

		var ar *asyncResult

		select {
		case ar = <-c.requestsChan:
		default:
			// Give the last chance for ready goroutines filling c.requestsChan :)
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case ar = <-c.requestsChan:
			case <-flushChan:
				if err = enc.flush(); err != nil {
					err = fmt.Errorf("cannot flush requests to: %s: %s", c.Addr, err)
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = getFlushChan(t, c.FlushDelay)
		}

		if ar.isCanceled() {
			if ar.reqData != nil {
				_ = ar.reqData.Close()
			}

			if ar.done != nil {
				ar.err = xrpc.ErrCanceled
				close(ar.done)
			} else {
				releaseAsyncResult(ar)
			}

			continue
		}

		if ar.done == nil {
			if ar.reqData != nil {
				_ = ar.reqData.Close()
			}
			releaseAsyncResult(ar)
			continue
		}

		pendingRequestsLock.Lock()
		n := len(pendingRequests)
		pendingRequests[msgID] = ar
		pendingRequestsLock.Unlock()

		if n > 10*c.PendingRequests {
			xlog.ErrorIDf(ar.reqid, "server: %s didn't return %d responses yet: closing connection", c.Addr, n)
			err = xrpc.ErrConnection
			return
		}

		header.method = ar.method
		header.msgID = msgID
		header.reqid = ar.reqid
		if ar.reqData != nil {
			header.bodySize = uint32(len(ar.reqData.Bytes()))
		} else {
			header.bodySize = 0
		}
		header.oid = ar.oid
		msg.header = header
		msg.body = ar.reqData

		if err = enc.encode(msg); err != nil {
			xlog.ErrorIDf(ar.reqid, "failed to send request to: %s: %s", c.Addr, err)
			return
		}
		msg.header = nil
		msg.body = nil
	}
}

func clientReader(c *Client, r net.Conn, pendingRequests map[uint64]*asyncResult, pendingRequestsLock *sync.Mutex, done chan<- error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				xlog.Errorf("panic when reading data from server: %s: %v", c.Addr, r)
			}
		}
		done <- err
	}()

	hash := xdigest.New()
	if c.encrypted {
		hash = nil
	}
	dec := newDecoder(r, c.RecvBufferSize, hash)
	msg := &message{
		header: new(respHeader),
	}
	for {
		err := dec.decode(msg)
		if err != nil {
			if err == xrpc.ErrTimeout {
				msg.body = nil
				continue // Keeping trying to read request header.
			}
			xlog.Errorf("failed to read request from %s: %s", r.RemoteAddr().String(), err)
			if msg.body != nil {
				_ = msg.body.Close()
			}
			return
		}

		body := msg.body
		h := msg.header.(*respHeader)
		msgID := h.msgID
		errno := h.errno
		n := h.bodySize

		pendingRequestsLock.Lock()
		ar, ok := pendingRequests[msgID]
		if ok {
			delete(pendingRequests, msgID)
		}
		pendingRequestsLock.Unlock()

		if !ok {
			xlog.Errorf("unexpected msgID: %d obtained from: %s", msgID, c.Addr)
			err = xrpc.ErrInternalServer
			if msg.body != nil {
				_ = msg.body.Close()
			}
			return
		}

		// Ignore response if any error. And the response must be nil.
		digest := binary.LittleEndian.Uint32(ar.oid[8:12])
		if !c.encrypted && n != 0 {
			actDigest := hash.Sum32()
			if actDigest != digest {
				xlog.ErrorID(ar.reqid, xerrors.WithMessage(xrpc.ErrChecksumMismatch, fmt.Sprintf("response exp: %d, but: %d", digest, actDigest)).Error())
				ar.err = xrpc.ErrChecksumMismatch
				if body != nil {
					_ = body.Close()
				}
				ar.respBody = nil
				msg.body = nil
				close(ar.done)
				hash.Reset()
				continue
			}
			hash.Reset()
		}

		if errno != 0 {
			ar.err = xrpc.Errno(errno).ToErr()
			if body != nil {
				_ = body.Close()
			}
			ar.respBody = nil
			msg.body = nil
			close(ar.done)
			continue
		}

		if n == 0 {
			if body != nil {
				_ = body.Close()
			}

			ar.respBody = nil
			msg.body = nil
			close(ar.done)
			continue
		}

		ar.respBody = body
		close(ar.done)
		msg.body = nil
	}
}

var asyncResultPool sync.Pool

func acquireAsyncResult() *asyncResult {
	v := asyncResultPool.Get()
	if v == nil {
		return &asyncResult{}
	}
	return v.(*asyncResult)
}

func releaseAsyncResult(ar *asyncResult) {
	ar.method = 0
	ar.reqid = 0
	ar.oid = [16]byte{}
	ar.reqData = nil

	ar.respBody = nil
	ar.done = nil
	ar.err = nil

	atomic.CompareAndSwapUint32(&ar.canceled, 1, 0)

	asyncResultPool.Put(ar)
}
