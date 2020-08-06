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
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/xerrors"

	"github.com/zaibyte/pkg/xdigest"

	"github.com/zaibyte/pkg/xrpc"

	"github.com/zaibyte/pkg/xlog"
)

// Server implements RPC server.
//
// Default server settings are optimized for high load, so don't override
// them without valid reason.
type Server struct {
	// Address to listen to for incoming connections.
	//
	// The address format depends on the underlying transport provided
	// by Server.Listener. The following transports are provided
	// out of the box:
	//   * TCP - see NewTCPServer() and NewTCPClient().
	//   * TLS (aka SSL) - see NewTLSServer() and NewTLSClient().
	//
	// By default TCP transport is used.
	Addr string

	// The maximum number of concurrent rpc calls the server may perform.
	// Default is DefaultConcurrency.
	Concurrency int

	// The maximum number of pending responses in the queue.
	// Default is DefaultPendingMessages.
	PendingResponses int

	// Size of send buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	SendBufferSize int

	// Size of recv buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	RecvBufferSize int

	// The maximum delay between response flushes to clients.
	//
	// Negative values lead to immediate requests' sending to the client
	// without their buffering. This minimizes rpc latency at the cost
	// of higher CPU and network usage.
	//
	// Default is DefaultFlushDelay.
	FlushDelay time.Duration

	// The server obtains new client connections via Listener.Accept().
	//
	// Override the Listener if you want custom underlying transport
	// and/or client authentication/authorization.
	// Don't forget overriding Client.Dial() callback accordingly.
	//
	// * NewTLSClient() and NewTLSServer() can be used for encrypted rpc.
	// * NewUnixClient() and NewUnixServer() can be used for fast local
	//   inter-process rpc.
	//
	// By default it returns TCP connections accepted from Server.Addr.
	Listener *defaultListener

	PutObj    xrpc.PutFunc
	GetObj    xrpc.GetFunc
	DeleteObj xrpc.DeleteFunc

	encrypted bool

	serverStopChan chan struct{}
	stopWg         sync.WaitGroup
}

const (
	// DefaultConcurrency is the default number of concurrent rpc calls
	// the server can process.
	DefaultConcurrency = 8 * 1024
	// DefaultServerSendBufferSize is the default size for Server send buffers.
	DefaultServerSendBufferSize = 64 * 1024
	// DefaultServerRecvBufferSize is the default size for Server receive buffers.
	DefaultServerRecvBufferSize = 64 * 1024
)

// Start starts rpc server.
func (s *Server) Start() error {

	if s.serverStopChan != nil {
		xlog.Panic("server is already running. Stop it before starting it again")
	}
	s.serverStopChan = make(chan struct{})

	if s.PutObj == nil || s.GetObj == nil || s.DeleteObj == nil {
		xlog.Panic("not enough handler function")
	}

	if s.Concurrency <= 0 {
		s.Concurrency = DefaultConcurrency
	}
	if s.PendingResponses <= 0 {
		s.PendingResponses = DefaultPendingMessages
	}
	if s.SendBufferSize <= 0 {
		s.SendBufferSize = DefaultServerSendBufferSize
	}
	if s.RecvBufferSize <= 0 {
		s.RecvBufferSize = DefaultServerRecvBufferSize
	}
	if s.FlushDelay == 0 {
		s.FlushDelay = DefaultFlushDelay
	}

	if s.Listener == nil {
		s.Listener = &defaultListener{}
	}
	if err := s.Listener.Init(s.Addr); err != nil {
		xlog.Errorf("cannot listen to: %s: %s", s.Addr, err.Error())
		return err
	}

	workersCh := make(chan struct{}, s.Concurrency)
	s.stopWg.Add(1)
	go serverHandler(s, workersCh)
	return nil
}

// Stop stops rpc server. Stopped server can be started again.
func (s *Server) Stop() {
	if s.serverStopChan == nil {
		xlog.Panic("server must be started before stopping it")
	}
	close(s.serverStopChan)
	s.stopWg.Wait()
	s.serverStopChan = nil
}

// Serve starts rpc server and blocks until it is stopped.
func (s *Server) Serve() error {
	if err := s.Start(); err != nil {
		return err
	}
	s.stopWg.Wait()
	return nil
}

func serverHandler(s *Server, workersCh chan struct{}) {
	defer s.stopWg.Done()

	var conn net.Conn
	var err error
	var stopping atomic.Value

	for {
		acceptChan := make(chan struct{})
		go func() {
			if conn, err = s.Listener.Accept(); err != nil {
				xlog.Errorf("failed to accept: %s", err.Error())
				if stopping.Load() == nil {
					xlog.Errorf("cannot accept new connection: %s", err)
				}
			}
			close(acceptChan)
		}()

		select {
		case <-s.serverStopChan:
			stopping.Store(true)
			_ = s.Listener.Close()
			<-acceptChan
			return
		case <-acceptChan:
		}

		if err != nil {
			select {
			case <-s.serverStopChan:
				return
			case <-time.After(time.Second):
			}
			continue
		}

		s.stopWg.Add(1)
		go serverHandleConnection(s, conn, workersCh)
	}
}

func serverHandleConnection(s *Server, conn net.Conn, workersCh chan struct{}) {
	defer s.stopWg.Done()

	var stopping atomic.Value
	var err error

	okHandshake := make(chan bool, 1)
	go func() {
		var buf [1]byte
		if _, err = conn.Read(buf[:]); err != nil {
			if stopping.Load() == nil {
				xlog.Errorf("failed to reading handshake from client: %s: %s", conn.RemoteAddr().String(), err)
			}
		}
		okHandshake <- buf[0] == 1
	}()

	select {
	case ok := <-okHandshake:
		if !ok || err != nil {
			_ = conn.Close()
			return
		}
	case <-s.serverStopChan:
		stopping.Store(true)
		_ = conn.Close()
		return
	case <-time.After(10 * time.Second):
		xlog.Errorf("cannot obtain handshake from client:%s during 10s", conn.RemoteAddr().String())
		_ = conn.Close()
		return
	}

	responsesChan := make(chan *serverMessage, s.PendingResponses)
	stopChan := make(chan struct{})

	readerDone := make(chan struct{})
	go serverReader(s, conn, responsesChan, stopChan, readerDone, workersCh)

	writerDone := make(chan struct{})
	go serverWriter(s, conn, responsesChan, stopChan, writerDone)

	select {
	case <-readerDone:
		close(stopChan)
		_ = conn.Close()
		<-writerDone
	case <-writerDone:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
	case <-s.serverStopChan:
		close(stopChan)
		_ = conn.Close()
		<-readerDone
		<-writerDone
	}
}

type serverMessage struct {
	method   uint8
	msgID    uint64
	reqid    uint64
	oid      [16]byte
	bodySize uint32
	reqbody  xrpc.Byteser

	resp xrpc.Byteser
	err  error
}

var serverMessagePool = &sync.Pool{
	New: func() interface{} {
		return &serverMessage{}
	},
}

func (s *serverMessage) reset() {
	s.method = 0
	s.msgID = 0
	s.reqid = 0
	s.reqbody = nil

	s.resp = nil
	s.err = nil
}

func serverReader(s *Server, r net.Conn, responsesChan chan<- *serverMessage,
	stopChan <-chan struct{}, done chan<- struct{}, workersCh chan struct{}) {

	defer func() {
		if r := recover(); r != nil {
			xlog.Errorf("panic when reading data from client: %v", r)
		}
		close(done)
	}()

	hash := xdigest.New()
	if s.encrypted {
		hash = nil
	}
	dec := newDecoder(r, s.RecvBufferSize, hash)
	req := &message{
		header: new(reqHeader),
	}
	for {
		err := dec.decode(req)
		if err != nil {
			if err == xrpc.ErrTimeout {
				continue // Keeping trying to read request header.
			}
			xlog.Errorf("failed to read request from %s: %s", r.RemoteAddr().String(), err)
			return
		}

		body := req.body
		req.body = nil

		h := req.header.(*reqHeader)
		m := serverMessagePool.Get().(*serverMessage)
		m.method = h.method
		m.msgID = h.msgID
		m.reqid = h.reqid
		m.oid = h.oid
		m.bodySize = h.bodySize

		digest := binary.LittleEndian.Uint32(m.oid[8:12])

		if !s.encrypted && m.bodySize != 0 {
			actDigest := hash.Sum32()
			if actDigest != digest {
				xlog.ErrorID(m.reqid, xerrors.WithMessage(xrpc.ErrChecksumMismatch, fmt.Sprintf("request exp: %d, but: %d", digest, actDigest)).Error())
				m.err = xrpc.ErrChecksumMismatch
				if body != nil {
					_ = body.Close()
				}
				m.reqbody = nil
			}
			hash.Reset()
		}
		if m.err == nil {
			m.reqbody = body
		}

		select {
		case workersCh <- struct{}{}:
		default:
			select {
			case workersCh <- struct{}{}:
			case <-stopChan:
				return
			}
		}

		// Haven read the request, handle request async, free the conn for the next request reading.
		go serveRequest(s, responsesChan, stopChan, m, workersCh)
	}
}

func serveRequest(s *Server, responsesChan chan<- *serverMessage, stopChan <-chan struct{}, m *serverMessage, workersCh <-chan struct{}) {

	if m.err == nil {
		reqBody := m.reqbody
		resp, err := callHandlerWithRecover(s, m.reqid, m.method, m.oid, reqBody)
		m.resp = resp
		m.err = err
	}

	if m.reqbody != nil {
		_ = m.reqbody.Close()
	}
	m.reqbody = nil

	// Select hack for better performance.
	// See https://github.com/valyala/gorpc/pull/1 for details.
	select {
	case responsesChan <- m:
	default:
		select {
		case responsesChan <- m:
		case <-stopChan:
		}
	}

	<-workersCh
}

func callHandlerWithRecover(s *Server, reqid uint64, method uint8, oid [16]byte, reqData xrpc.Byteser) (resp xrpc.Byteser, err error) {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			err = fmt.Errorf("panic occured: %v\nStack trace: %s", x, stackTrace[:n])
			xlog.ErrorID(reqid, err.Error())
		}
	}()

	switch method {
	case objPutMethod:
		err = s.PutObj(reqid, oid, reqData)
	case objGetMethod:
		resp, err = s.GetObj(reqid, oid)
	case objDelMethod:
		err = s.DeleteObj(reqid, oid)
	default:
		err = xrpc.ErrNotImplemented
	}

	return
}

func isServerStop(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

func serverWriter(s *Server, w net.Conn, responsesChan <-chan *serverMessage, stopChan <-chan struct{}, done chan<- struct{}) {
	defer func() { close(done) }()

	t := time.NewTimer(s.FlushDelay)
	var flushChan <-chan time.Time
	enc := newEncoder(w, s.SendBufferSize)
	msg := new(message)
	header := new(respHeader)
	for {
		var m *serverMessage

		select {
		case m = <-responsesChan:
		default:
			// Give the last chance for ready goroutines filling responsesChan :)
			runtime.Gosched()

			select {
			case <-stopChan:
				return
			case m = <-responsesChan:
			case <-flushChan:
				if err := enc.flush(); err != nil {
					if !isServerStop(stopChan) {
						xlog.Errorf("cannot flush requests to: %s: %s", w.RemoteAddr().String(), err)
					}
					return
				}
				flushChan = nil
				continue
			}
		}

		if flushChan == nil {
			flushChan = getFlushChan(t, s.FlushDelay)
		}

		resp := m.resp
		reqid := m.reqid

		header.msgID = m.msgID
		if resp != nil {
			header.bodySize = uint32(len(resp.Bytes()))
		} else {
			header.bodySize = 0
		}
		header.errno = uint16(xrpc.ErrToErrno(m.err))
		msg.header = header
		msg.body = resp

		m.reset()
		serverMessagePool.Put(m)

		if err := enc.encode(msg); err != nil {
			xlog.ErrorIDf(reqid, "failed to send response to: %s: %s", w.RemoteAddr().String(), err)
			return
		}

		msg.body = nil
	}
}
