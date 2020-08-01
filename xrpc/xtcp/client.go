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
	"hash/crc32"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

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
	//
	// The address format depends on the underlying transport provided
	// by Client.Dial. The following transports are provided out of the box:
	//   * TCP - see NewTCPClient() and NewTCPServer().
	//   * TLS - see NewTLSClient() and NewTLSServer().
	//
	// If created by NewClient(), it will chose transport automatically.
	Addr string

	// The number of concurrent connections the client should establish
	// to the sever.
	// Default is DefaultClientConns.
	Conns int

	// The maximum number of pending requests in the queue.
	//
	// The number of pending requests should exceed the expected number
	// of concurrent goroutines calling client's methods.
	// Otherwise a lot of ClientError.Overflow errors may appear.
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

	// Size of Client write payload in bytes.
	// Default value is DefaultClientSendPayloadSize.
	SendPayloadSize int

	// Size of Client read payload in bytes.
	// Default value is DefaultClientRecvPayloadSize.
	RecvPayloadSize int

	// The client calls this callback when it needs new connection
	// to the server.
	// The client passes Client.Addr into Dial().
	//
	// Override this callback if you want custom underlying transport
	// and/or authentication/authorization.
	// Don't forget overriding Server.Listener accordingly.
	//
	//
	// * NewTLSClient() and NewTLSServer() can be used for encrypted rpc.
	// * NewUnixClient() and NewUnixServer() can be used for fast local
	//   inter-process rpc.
	//
	// By default it returns TCP connections established to the Client.Addr.
	Dial DialFunc

	encrypted bool

	router *Router

	requestsChan chan *AsyncResult

	stopChan chan struct{}
	stopWg   sync.WaitGroup
}

// AsyncResult is a result returned from Client.CallAsync().
type AsyncResult struct {
	Method uint16
	ReqID  uint64
	// The response can be read only after <-Done unblocks.
	Response xrpc.Marshaler

	// Response is become available after <-Done unblocks.
	Done <-chan struct{}

	// The error can be read only after <-Done unblocks.
	Error error

	request  xrpc.Marshaler
	extraReq *xrpc.Buffer
	done     chan struct{}
	canceled uint32
}

// Cancel cancels async call.
//
// Canceled call isn't sent to the server unless it is already sent there.
// Canceled call may successfully complete if it has been already sent
// to the server before Cancel call.
//
// It is safe calling this function multiple times from concurrently
// running goroutines.
func (r *AsyncResult) Cancel() {
	atomic.StoreUint32(&r.canceled, 1)
}

func (r *AsyncResult) isCanceled() bool {
	return atomic.LoadUint32(&r.canceled) != 0
}

const (
	// DefaultRequestTimeout is the default timeout for client request.
	DefaultRequestTimeout = 5 * time.Second

	// DefaultClientSendBufferSize is the default size for Client send buffers.
	DefaultClientSendBufferSize = 64 * 1024

	// DefaultClientRecvBufferSize is the default size for Client receive buffers.
	DefaultClientRecvBufferSize = 64 * 1024

	// DefaultClientSendPayloadSize is the default size for Client write payload buffers.
	// In Zai, raw binary will be put in extra request, so it could be small.
	DefaultClientSendPayloadSize = 1024

	// DefaultClientRecvPayloadSize is the default size for Client read payload buffers.
	DefaultClientRecvPayloadSize = 128 * 1024

	// DefaultClientConns is the default connection numbers for Client.
	DefaultClientConns = 4
)

// Start starts rpc client. Establishes connection to the server on Client.Addr.
func (c *Client) Start() {

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

	if c.SendPayloadSize <= 0 {
		c.SendPayloadSize = DefaultClientSendPayloadSize
	}

	if c.RecvPayloadSize <= 0 {
		c.RecvPayloadSize = DefaultClientRecvPayloadSize
	}

	c.requestsChan = make(chan *AsyncResult, c.PendingRequests)
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
}

// Stop stops rpc client. Stopped client can be started again.
func (c *Client) Stop() {
	if c.stopChan == nil {
		xlog.Panic("client must be started before stopping it")
	}
	close(c.stopChan)
	c.stopWg.Wait()
	c.stopChan = nil
}

// Call sends the given request to the server and obtains response
// from the server.
// Returns non-nil error if the response cannot be obtained during
// Client.RequestTimeout or server connection problems occur.
//
// Hint: use Router for distinct calls' construction.
//
// Don't forget starting the client with Client.Start() before calling Client.Call().
func (c *Client) Call(reqid uint64, method uint8, request xrpc.Marshaler, extraReq *xrpc.Buffer, resp xrpc.Marshaler) (err error) {
	return c.CallTimeout(reqid, method, request, extraReq, c.RequestTimeout, resp)
}

// CallTimeout sends the given request to the server and obtains response
// from the server.
// Returns non-nil error if the response cannot be obtained during
// Client.RequestTimeout or server connection problems occur.
//
// Hint: use Router for distinct calls' construction.
//
// Don't forget starting the client with Client.Start() before calling Client.Call().
func (c *Client) CallTimeout(reqid uint64, method uint8, request xrpc.Marshaler, extraReq *xrpc.Buffer, timeout time.Duration, resp xrpc.Marshaler) (err error) {

	var ar *AsyncResult
	if ar, err = c.callAsync(reqid, method, request, extraReq, resp); err != nil {
		return err
	}

	t := acquireTimer(timeout)

	select {
	case <-ar.Done:
		err = ar.Error
		releaseAsyncResult(ar)
	case <-t.C:
		ar.Cancel()
		err = xrpc.ErrTimeout
	}

	releaseTimer(t)
	return
}

func (c *Client) callAsync(reqid uint64, method uint8, request xrpc.Marshaler, extraReq *xrpc.Buffer, resp xrpc.Marshaler) (ar *AsyncResult, err error) {

	if reqid == 0 {
		reqid = uid.MakeReqID()
	}

	if c.router.handlers[method] == nil {
		return nil, xrpc.ErrNotImplemented
	}

	ar = acquireAsyncResult()
	ar.ReqID = reqid
	ar.Method = uint16(method)
	ar.request = request
	ar.Response = resp
	ar.extraReq = extraReq
	ar.done = make(chan struct{})
	ar.Done = ar.done

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
				ar2.Error = xrpc.ErrRequestQueueOverflow
				close(ar2.done)
			} else {
				releaseAsyncResult(ar2)
			}
		default:
		}

		select {
		case c.requestsChan <- ar:
			return ar, nil
		default:
			// Release ar even if use pool, since ar wasn't exposed
			// to the caller yet.
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
			case <-time.After(time.Second):
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

	var err error
	stopChan := make(chan struct{})

	pendingRequests := make(map[uint64]*AsyncResult)
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
		ar.Error = err
		close(ar.done)
	}
}

func clientWriter(c *Client, w net.Conn,
	pendingRequests map[uint64]*AsyncResult, pendingRequestsLock *sync.Mutex,
	stopChan <-chan struct{}, done chan<- error) {

	var err error
	defer func() { done <- err }()

	payload := make([]byte, c.SendPayloadSize)
	headerBuf := make([]byte, requestHeaderSize)

	var msgID uint64 = 1
	for {
		var ar *AsyncResult

		select {
		case ar = <-c.requestsChan:
		default:
			// Give the last chance for ready goroutines filling c.requestsChan :)
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case ar = <-c.requestsChan:
			}
		}

		if ar.isCanceled() {
			if ar.done != nil {
				ar.Error = xrpc.ErrCanceled
				close(ar.done)
			} else {
				releaseAsyncResult(ar)
			}

			continue
		}

		msgID++
		pendingRequestsLock.Lock()
		n := len(pendingRequests)
		pendingRequests[msgID] = ar
		pendingRequestsLock.Unlock()

		if n > 10*c.PendingRequests {
			xlog.ErrorIDf(ar.ReqID, "server: %s didn't return %d responses yet: closing connection", c.Addr, n)
			err = xrpc.ErrConnection
			return
		}

		err = sendRequest(c, w, ar, msgID, headerBuf, payload)
		if err != nil {
			return
		}
	}
}

func sendRequest(c *Client, w net.Conn, ar *AsyncResult, msgID uint64, headerBuf, payload []byte) (err error) {

	var n int
	if ar.request != nil {
		n, err = ar.request.MarshalLen()
		if err != nil {
			xlog.ErrorIDf(ar.ReqID, "request get marshal len failed: %s", err.Error())
			err = xrpc.ErrBadRequest
			return
		}
	}
	var extraN int
	if ar.extraReq != nil {
		extraN = ar.extraReq.Len()
	}

	h := &requestHeader{
		method:    ar.Method,
		msgID:     msgID,
		reqSize:   uint32(n),
		extraSize: uint32(extraN),
		reqid:     ar.ReqID,
	}

	var buf []byte

	if ar.request != nil {
		if len(payload) < n {
			buf = make([]byte, n)
		} else {
			buf = payload[:n]
		}
		buf, err = ar.request.MarshalTo(buf)
		if err != nil {
			xlog.ErrorIDf(ar.ReqID, "request marshal failed: %s", err.Error())
			err = xrpc.ErrInternalServer
			return
		}
	}

	if !c.encrypted {
		crc := crc32.New(xdigest.CrcTbl)
		crc.Write(buf)
		if ar.extraReq != nil {
			crc.Write(ar.extraReq.Bytes())
		}
		h.crc = crc.Sum32()
	}
	h.encode(headerBuf)

	tt := time.Now().Add(magicNumberDuration).Add(headerDuration)
	if err = w.SetWriteDeadline(tt); err != nil {
		xlog.ErrorIDf(h.reqid, "failed to set magic number & header write deadline to %s: %s", c.Addr, err)
		err = xrpc.ErrConnection
		return
	}
	if _, err = w.Write(magicNumber[:]); err != nil {
		xlog.ErrorIDf(h.reqid, "failed to write magic number to %s: %s", c.Addr, err)
		err = xrpc.ErrConnection
		return
	}

	_, err = w.Write(headerBuf)
	if err != nil {
		xlog.ErrorIDf(h.reqid, "failed to write header to %s: %s", c.Addr, err)
		err = xrpc.ErrConnection
		return
	}

	if n != 0 {
		err = sendBytes(ar.ReqID, c.Addr, w, buf, c.SendBufferSize)
		if err != nil {
			err = xrpc.ErrConnection
			return
		}
	}

	if extraN != 0 {
		err = sendBytes(ar.ReqID, c.Addr, w, ar.extraReq.Bytes(), c.SendBufferSize)
		if err != nil {
			err = xrpc.ErrConnection
			return
		}
	}

	return nil
}

func sendBytes(reqid uint64, addr string, w net.Conn, buf []byte, bufSize int) (err error) {

	sent := 0
	var tt time.Time
	for sent < len(buf) {
		if sent+bufSize > len(buf) {
			bufSize = len(buf) - sent
		}
		tt = time.Now().Add(writeDuration)
		if err = w.SetWriteDeadline(tt); err != nil {
			xlog.ErrorIDf(reqid, "failed to set write deadline to %s: %s", addr, err)
			return
		}
		if _, err = w.Write(buf[sent : sent+bufSize]); err != nil {
			xlog.ErrorIDf(reqid, "failed to write to %s: %s", addr, err)
			return
		}
		sent += bufSize
	}
	if sent != len(buf) {
		xlog.ErrorIDf(reqid, "unexpected sent size to %s: %d, buf want %d", addr, sent, len(buf))
		return xrpc.ErrConnection
	}
	return nil
}

func clientReader(c *Client, r net.Conn, pendingRequests map[uint64]*AsyncResult, pendingRequestsLock *sync.Mutex, done chan<- error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				xlog.Errorf("panic when reading data from server: %s: %v", c.Addr, r)
			}
		}
		done <- err
	}()

	magicNum := make([]byte, len(magicNumber))
	payload := make([]byte, c.RecvPayloadSize)
	headerBuf := make([]byte, respHeaderSize)

	for {
		err = readMagicNumber(r, magicNum)
		if err != nil {
			if err == xrpc.ErrTimeout {
				continue // Keeping trying to read response.
			}
			xlog.Errorf("failed to read magic number from %s: %s", c.Addr, err)
			return
		}

		header, err2 := readRespHeader(r, headerBuf)
		if err2 != nil {
			err = err2
			xlog.Errorf("failed to read header from %s: %s", c.Addr, err)
			return
		}

		pendingRequestsLock.Lock()
		ar, ok := pendingRequests[header.msgID]
		if ok {
			delete(pendingRequests, header.msgID)
		}
		pendingRequestsLock.Unlock()

		if !ok {
			xlog.ErrorIDf(ar.ReqID, "unexpected msgID: %d obtained from: %s", header.msgID, c.Addr)
			err = xrpc.ErrInternalServer
			return
		}

		if header.errno != 0 {
			ar.Error = xrpc.Errno(header.errno)
			close(ar.done)
			continue // Ignore response if any error. And the response must be nil.
		}

		isNop := c.router.isNopResp(uint8(ar.Method))

		n := header.respSize
		if n == 0 {
			if !isNop {
				xlog.ErrorID(ar.ReqID, "invalid response length")
				err = xrpc.ErrInternalServer
				ar.Error = err
				close(ar.done)
				return
			} else {
				close(ar.done) // NopResp. Returns a nil response.
				continue
			}
		}

		var buf []byte
		resp := ar.Response
		if c.router.isBytesResp(uint8(ar.Method)) {
			buf, _ = resp.MarshalTo(nil) // No side effect when it's a *xprc.Buffer.
		} else {
			if header.respSize > uint32(len(payload)) {
				buf = make([]byte, n)
			} else {
				buf = payload[:n]
			}
		}

		received := uint32(0)
		var recvBuf []byte
		if n < uint32(c.RecvBufferSize) {
			recvBuf = buf[:n]
		} else {
			recvBuf = buf[:c.RecvBufferSize]
		}
		toRead := n
		tt := time.Now()
		crc := crc32.New(xdigest.CrcTbl)
		for toRead > 0 {
			tt = tt.Add(readDuration)
			if err = r.SetReadDeadline(tt); err != nil {
				xlog.ErrorIDf(ar.ReqID, "failed to set read deadline: %s, %s", c.Addr, err.Error())
				err = xrpc.ErrConnection
				ar.Error = err
				close(ar.done)
				return
			}
			if _, err = io.ReadFull(r, recvBuf); err != nil {
				xlog.ErrorIDf(ar.ReqID, "failed to read: %s, %s", c.Addr, err.Error())
				err = xrpc.ErrConnection
				ar.Error = err
				close(ar.done)
				return
			}
			if !c.encrypted {
				crc.Write(recvBuf)
			}
			toRead -= uint32(len(recvBuf))
			received += uint32(len(recvBuf))
			if toRead < uint32(c.RecvBufferSize) {
				recvBuf = buf[received : received+toRead]
			} else {
				recvBuf = buf[received : received+uint32(c.RecvBufferSize)]
			}
		}
		if received != n {
			xlog.ErrorIDf(ar.ReqID, "unexpected received size: %d, but want %d", received, n)
			err = xrpc.ErrInternalServer
			ar.Error = err
			close(ar.done)
			return
		}
		if !c.encrypted && crc.Sum32() != header.crc {
			xlog.ErrorID(ar.ReqID, "invalid checksum")
			err = xrpc.ErrChecksumMismatch
			ar.Error = err
			close(ar.done)
			return
		}

		err = resp.UnmarshalBinary(buf)
		if err != nil {
			xlog.ErrorIDf(ar.ReqID, "cannot decode response: %s", err.Error())
			err = xrpc.ErrInternalServer
			ar.Error = err
			close(ar.done)
			return
		}

		close(ar.done)
	}
}

func readRespHeader(r net.Conn, headerBuf []byte) (header *respHeader, err error) {
	tt := time.Now().Add(headerDuration)
	if err = r.SetReadDeadline(tt); err != nil {
		err = xrpc.ErrConnection
		return
	}
	if _, err = io.ReadFull(r, headerBuf); err != nil {
		err = xrpc.ErrConnection
		return
	}

	header = new(respHeader)
	if err = header.decode(headerBuf); err != nil {
		return
	}
	return
}

var asyncResultPool sync.Pool

func acquireAsyncResult() *AsyncResult {
	v := asyncResultPool.Get()
	if v == nil {
		return &AsyncResult{}
	}
	return v.(*AsyncResult)
}

func releaseAsyncResult(ar *AsyncResult) {
	ar.Method = 0
	ar.ReqID = 0
	ar.Response = nil
	ar.Done = nil
	ar.Error = nil
	ar.request = nil
	ar.done = nil
	asyncResultPool.Put(ar)
}
