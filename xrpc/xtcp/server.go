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
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/xdigest"

	"github.com/zaibyte/pkg/xrpc"

	"github.com/zaibyte/pkg/xlog"
)

// HandlerFunc is a server handler function.
//
// req is the combination of mainReq & extReq in Client.Call(),
// mreqLen is the mainReq's length in bytes.
//
// HandlerFunc will be invoked by Server for handling request.
// Before this process, the request & extraRequest will be read from the net connection,
// (after reading, the connection can be reused for the next reading)
// using *xprc.Buffer could reducing the GC overhead by sync.Pool.
//
// After resp written into connection, it will be freed.
type HandlerFunc func(reqid uint64, req *xrpc.Buffer, mreqLen int) (resp xrpc.MarshalFreer, err error)

const (
	// DefaultConcurrency is the default number of concurrent rpc calls
	// the server can process.
	DefaultConcurrency = 8 * 1024

	// DefaultServerSendPayloadSize is the default size for Client write payload buffers.
	DefaultServerSendPayloadSize = 512 * 1024

	// DefaultServerRecvPayloadSize is the default size for Client read payload buffers.
	DefaultServerRecvPayloadSize = 1024

	// DefaultServerSendBufferSize is the default size for Server send buffers.
	DefaultServerSendBufferSize = 64 * 1024
	// DefaultServerRecvBufferSize is the default size for Server receive buffers.
	DefaultServerRecvBufferSize = 64 * 1024
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

	// Router function for incoming requests.
	//
	// Server calls this function for each incoming request.
	// The function must process the request and return the corresponding response.
	//
	// Hint: use Router for HandlerFunc construction.
	Router *Router

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

	// Size of Client write payload in bytes.
	// Default value is DefaultServerSendPayloadSize.
	SendPayloadSize int

	// Size of Client read payload in bytes.
	// Default value is DefaultServerRecvPayloadSize.
	RecvPayloadSize int

	// The server obtains new client connections via Listener.Accept().
	//
	// Override the listener if you want custom underlying transport
	// and/or client authentication/authorization.
	// Don't forget overriding Client.Dial() callback accordingly.
	//
	// * NewTLSClient() and NewTLSServer() can be used for encrypted rpc.
	// * NewUnixClient() and NewUnixServer() can be used for fast local
	//   inter-process rpc.
	//
	// By default it returns TCP connections accepted from Server.Addr.
	Listener Listener

	encrypted bool

	serverStopChan chan struct{}
	stopWg         sync.WaitGroup
}

// Start starts rpc server.
func (s *Server) Start() error {

	if s.Router == nil {
		xlog.Panic("server.Router cannot be nil")
	}

	if s.serverStopChan != nil {
		xlog.Panic("server is already running. Stop it before starting it again")
	}
	s.serverStopChan = make(chan struct{})

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
	if s.SendPayloadSize <= 0 {
		s.SendPayloadSize = DefaultServerSendPayloadSize
	}
	if s.RecvPayloadSize <= 0 {
		s.RecvPayloadSize = DefaultServerRecvPayloadSize
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
	Method   uint16
	MsgID    uint64
	ReqID    uint64
	Request  xrpc.Marshaler
	ExtraReq *xrpc.Buffer
	Response xrpc.Marshaler
	Error    error
}

var serverMessagePool = &sync.Pool{
	New: func() interface{} {
		return &serverMessage{}
	},
}

func (s *serverMessage) reset() {
	s.Method = 0
	s.MsgID = 0
	s.ReqID = 0
	s.Request = nil
	if s.ExtraReq != nil {
		s.ExtraReq.Free()
	}
	s.ExtraReq = nil
	s.Response = nil
	s.Error = nil
}

func serverReader(s *Server, r net.Conn, responsesChan chan<- *serverMessage,
	stopChan <-chan struct{}, done chan<- struct{}, workersCh chan struct{}) {

	defer func() {
		if r := recover(); r != nil {
			xlog.Errorf("panic when reading data from client: %v", r)
		}
		close(done)
	}()

	magicNum := make([]byte, len(magicNumber))
	payload := make([]byte, s.RecvPayloadSize) // TODO can't use payload here, it will be polluted.
	headerBuf := make([]byte, requestHeaderSize)

	for {
		err := readMagicNumber(r, magicNum)
		if err != nil {
			if err == xrpc.ErrTimeout {
				continue // Keeping trying to read request.
			}
			xlog.Errorf("failed to read magic number from %s: %s", r.RemoteAddr().String(), err)
			return
		}

		header, err := readReqHeader(r, headerBuf)
		if err != nil {
			xlog.Errorf("failed to read header from %s: %s", r.RemoteAddr().String(), err)
			return
		}

		m := serverMessagePool.Get().(*serverMessage)
		m.Method = header.method
		m.MsgID = header.msgID
		m.ReqID = header.reqid

		n := header.reqSize

		var buf []byte
		if n > uint32(len(payload)) {
			buf = make([]byte, n) // TODO use a small byte slice pool?
		} else {
			buf = payload[:n] // TODO can't use payload here, because it'll be handled in another goroutine.
		}

		crc := crc32.New(xdigest.CrcTbl)

		err = readBytes(m.ReqID, r, buf, s.RecvBufferSize, s.encrypted, crc)
		if err != nil {
			m.reset()
			serverMessagePool.Put(m)
			return
		}
		reqT := s.Router.reqTypes[m.Method]                       // TODO handle handles byte slice maybe a better idea?
		req, ok := reflect.New(reqT).Interface().(xrpc.Marshaler) // TODO use pool to reducing GC.
		if !ok {
			xlog.ErrorID(m.ReqID, "invalid request type, not Marshaler")
			m.reset()
			serverMessagePool.Put(m)
			return
		}
		err = req.UnmarshalBinary(buf)
		if err != nil {
			xlog.ErrorIDf(m.ReqID, "cannot decode request: %s", err.Error())
			m.reset()
			serverMessagePool.Put(m)
			return
		}
		m.Request = req

		if header.extraSize != 0 {
			ext := xrpc.GetBytes()
			err = readBytes(m.ReqID, r, ext.Bytes(), s.RecvBufferSize, s.encrypted, crc)
			if err != nil {
				m.reset()
				serverMessagePool.Put(m)
				return
			}
			m.ExtraReq = ext
		}

		if !s.encrypted {
			incoming := crc.Sum32()
			if incoming != header.crc {
				xlog.ErrorID(m.ReqID, "failed to read request: invalid checksum")
				m.reset()
				serverMessagePool.Put(m)
				return
			}
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

func readReqHeader(r net.Conn, headerBuf []byte) (header *requestHeader, err error) {
	tt := time.Now().Add(headerDuration)
	if err = r.SetReadDeadline(tt); err != nil {
		err = xrpc.ErrConnection
		return
	}
	if _, err = io.ReadFull(r, headerBuf); err != nil {
		err = xrpc.ErrConnection
		return
	}

	header = new(requestHeader)
	if err = header.decode(headerBuf); err != nil {
		return
	}
	return
}

func readBytes(reqid uint64, r net.Conn, buf []byte, bufferSize int, encrypted bool, crc hash.Hash32) (err error) {

	n := len(buf)
	received := 0
	var recvBuf []byte
	if n < bufferSize {
		recvBuf = buf[:n]
	} else {
		recvBuf = buf[:bufferSize]
	}
	toRead := n
	tt := time.Now()
	for toRead > 0 {
		tt = tt.Add(readDuration)
		if err = r.SetReadDeadline(tt); err != nil {
			xlog.ErrorIDf(reqid, "failed to set read deadline: %s, %s", r.RemoteAddr().String(), err.Error())
			return
		}
		if _, err = io.ReadFull(r, recvBuf); err != nil {
			xlog.ErrorIDf(reqid, "failed to read: %s, %s", r.RemoteAddr().String(), err.Error())
			return
		}
		if !encrypted {
			crc.Write(recvBuf)
		}
		toRead -= len(recvBuf)
		received += len(recvBuf)
		if toRead < bufferSize {
			recvBuf = buf[received : received+toRead]
		} else {
			recvBuf = buf[received : received+bufferSize]
		}
	}
	if received != n {
		xlog.ErrorIDf(reqid, "unexpected received size: %d, but want %d", received, n)
		return xrpc.ErrConnection
	}
	return
}

func serveRequest(s *Server, responsesChan chan<- *serverMessage, stopChan <-chan struct{}, m *serverMessage, workersCh <-chan struct{}) {
	request := m.Request
	ext := m.ExtraReq
	m.Request = nil
	m.ExtraReq = nil

	response, err := callHandlerWithRecover(s.Router, m.ReqID, m.Method, request, ext)

	m.ExtraReq.Free()
	m.Response = response
	m.Error = err

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

func callHandlerWithRecover(router *Router, reqid uint64, method uint16, request xrpc.Marshaler, ext *xrpc.Buffer) (response xrpc.Marshaler, err error) {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			err = fmt.Errorf("panic occured: %v\nStack trace: %s", x, stackTrace[:n])
			xlog.ErrorID(reqid, err.Error())
		}
	}()
	return router.Handle(reqid, uint8(method), request, ext)
}

func serverWriter(s *Server, w net.Conn, responsesChan <-chan *serverMessage, stopChan <-chan struct{}, done chan<- struct{}) {
	defer func() { close(done) }()

	e := newMessageEncoder(w, s.SendBufferSize)
	defer e.Close()

	var wr wireResponse
	headerBuf := make([]byte, respHeaderSize)
	payload := make([]byte, s.SendPayloadSize)
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
			}

			wr.ID = m.MsgID
			wr.Response = m.Response
			wr.Error = m.Error

			m.Response = nil
			m.Error = ""
			serverMessagePool.Put(m)

			if err := e.Encode(wr); err != nil {
				s.LogError("xtcp.Server: Cannot send response to wire: [%s]", err)
				return
			}
			wr.Response = nil
			wr.Error = ""

		}
	}
}

func sendResp(s *Server, w net.Conn, m *serverMessage, headerBuf, payload []byte) (err error) {

	n, err := m.Response.MarshalLen()
	if err != nil {
		xlog.ErrorIDf(m.ReqID, "request get marshal len failed: %s", err.Error())
		err = xrpc.ErrBadRequest
		return
	}

	// TODO if n == 0, it must be nop, if not return error

	h := &respHeader{
		msgID:    m.MsgID,
		errno:    uint16(xrpc.ErrToErrno(m.Error)),
		respSize: uint32(n),
	}

	var buf []byte
	if s.Router.isBytesResp(uint8(m.Method)) {
		// TODO how to reduce GC?
	}
	if len(payload) < n {
		buf = make([]byte, n)
	} else {
		buf = payload[:n]
	}

	err = m.Request.MarshalTo(buf)
	if err != nil {
		xlog.ErrorIDf(m.ReqID, "request marshal failed: %s", err.Error())
		err = xrpc.ErrInternalServer
		return
	}

	if !s.encrypted {
		h.crc = xdigest.Checksum(buf)
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

	err = sendBytes(m.ReqID, c.Addr, w, buf, c.SendBufferSize)
	if err != nil {
		err = xrpc.ErrConnection
		return
	}

	err = sendBytes(m.ReqID, c.Addr, w, m.extraReq, c.SendBufferSize)
	if err != nil {
		err = xrpc.ErrConnection
		return
	}

	return nil
}
