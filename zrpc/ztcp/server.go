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

package ztcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// HandlerFunc is a server handler function.
//
// Request and response types may be arbitrary.
// All the request and response types the HandlerFunc may use must be registered
// with RegisterType() before starting the server.
// There is no need in registering base Go types such as int, string, bool,
// float64, etc. or arrays, slices and maps containing base Go types.
//
// Hint: use Dispatcher for HandlerFunc construction.
type HandlerFunc func(request interface{}) (response interface{})

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
	//   * Unix sockets - see NewUnixServer() and NewUnixClient().
	//
	// By default TCP transport is used.
	Addr string

	// Handler function for incoming requests.
	//
	// Server calls this function for each incoming request.
	// The function must process the request and return the corresponding response.
	//
	// Hint: use Dispatcher for HandlerFunc construction.
	Handler HandlerFunc

	// The maximum number of concurrent rpc calls the server may perform.
	// Default is DefaultConcurrency.
	Concurrency int

	// The maximum delay between response flushes to clients.
	//
	// Negative values lead to immediate requests' sending to the client
	// without their buffering. This minimizes rpc latency at the cost
	// of higher CPU and network usage.
	//
	// Default is DefaultFlushDelay.
	FlushDelay time.Duration

	// The maximum number of pending responses in the queue.
	// Default is DefaultPendingMessages.
	PendingResponses int

	// Size of send buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	SendBufferSize int

	// Size of recv buffer per each underlying connection in bytes.
	// Default is DefaultBufferSize.
	RecvBufferSize int

	// The server obtains new client connections via Listener.Accept().
	//
	// Override the listener if you want custom underlying transport
	// and/or client authentication/authorization.
	// Don't forget overriding Client.Dial() callback accordingly.
	//
	// See also OnConnect for authentication/authorization purposes.
	//
	// * NewTLSClient() and NewTLSServer() can be used for encrypted rpc.
	// * NewUnixClient() and NewUnixServer() can be used for fast local
	//   inter-process rpc.
	//
	// By default it returns TCP connections accepted from Server.Addr.
	Listener Listener

	// LogError is used for error logging.
	//
	// By default the function set via SetErrorLogger() is used.
	LogError LoggerFunc

	serverStopChan chan struct{}
	stopWg         sync.WaitGroup
}

// Start starts rpc server.
//
// All the request and response types the Handler may use must be registered
// with RegisterType() before starting the server.
// There is no need in registering base Go types such as int, string, bool,
// float64, etc. or arrays, slices and maps containing base Go types.
func (s *Server) Start() error {
	if s.LogError == nil {
		s.LogError = errorLogger
	}
	if s.Handler == nil {
		panic("ztcp.Server: Server.Handler cannot be nil")
	}

	if s.serverStopChan != nil {
		panic("ztcp.Server: server is already running. Stop it before starting it again")
	}
	s.serverStopChan = make(chan struct{})

	if s.Concurrency <= 0 {
		s.Concurrency = DefaultConcurrency
	}
	if s.FlushDelay == 0 {
		s.FlushDelay = DefaultFlushDelay
	}
	if s.PendingResponses <= 0 {
		s.PendingResponses = DefaultPendingMessages
	}
	if s.SendBufferSize <= 0 {
		s.SendBufferSize = DefaultBufferSize
	}
	if s.RecvBufferSize <= 0 {
		s.RecvBufferSize = DefaultBufferSize
	}

	if s.Listener == nil {
		s.Listener = &defaultListener{}
	}
	if err := s.Listener.Init(s.Addr); err != nil {
		err = fmt.Errorf("ztcp.Server: [%s]. Cannot listen to: [%s]", s.Addr, err)
		s.LogError("%s", err)
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
		panic("ztcp.Server: server must be started before stopping it")
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
					s.LogError("ztcp.Server: [%s]. Cannot accept new connection: [%s]", s.Addr, err)
				}
			}
			close(acceptChan)
		}()

		select {
		case <-s.serverStopChan:
			stopping.Store(true)
			s.Listener.Close()
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

	var err error
	var stopping atomic.Value

	zChan := make(chan struct{}, 1)
	go func() {
		var buf [2]byte
		err = readMagicNumber(conn, buf[:])
		if err != nil {
			if stopping.Load() == nil {
				s.LogError("ztcp.Server: Error when reading magic number from client: [%s]", err)
			}
		}

		zChan <- struct{}{}
	}()
	select {
	case <-zChan:
		if err != nil {
			conn.Close()
			return
		}
	case <-s.serverStopChan:
		stopping.Store(true)
		conn.Close()
		return
	case <-time.After(2 * time.Second):
		s.LogError("ztcp.Server: Cannot obtain reading magic number from client during 2s")
		conn.Close()
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
		conn.Close()
		<-writerDone
	case <-writerDone:
		close(stopChan)
		conn.Close()
		<-readerDone
	case <-s.serverStopChan:
		close(stopChan)
		conn.Close()
		<-readerDone
		<-writerDone
	}
}

func readMagicNumber(conn net.Conn, magicNum []byte) error {

	tt := time.Now().Add(magicNumberDuration)
	if err := conn.SetReadDeadline(tt); err != nil {
		return err
	}

	if _, err := io.ReadFull(conn, magicNum); err != nil {
		return err
	}
	if bytes.Equal(magicNum, poisonNumber[:]) {
		return ErrPoisonReceived
	}
	if !bytes.Equal(magicNum, magicNumber[:]) {
		return ErrBadMessage
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}
	return nil
}

type serverMessage struct {
	ID       uint64
	Request  interface{}
	Response interface{}
	Error    string
}

var serverMessagePool = &sync.Pool{
	New: func() interface{} {
		return &serverMessage{}
	},
}

func isClientDisconnect(err error) bool {
	return err == io.ErrUnexpectedEOF || err == io.EOF
}

func isServerStop(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

func serverReader(s *Server, r io.Reader, responsesChan chan<- *serverMessage,
	stopChan <-chan struct{}, done chan<- struct{}, workersCh chan struct{}) {

	defer func() {
		if r := recover(); r != nil {
			s.LogError("ztcp.Server: Panic when reading data from client: %v", r)
		}
		close(done)
	}()

	d := newMessageDecoder(r, s.RecvBufferSize)
	defer d.Close()

	var wr wireRequest
	for {
		if err := d.Decode(&wr); err != nil {
			if !isClientDisconnect(err) && !isServerStop(stopChan) {
				s.LogError("ztcp.Server: Cannot decode request: [%s]", err)
			}
			return
		}

		m := serverMessagePool.Get().(*serverMessage)
		m.ID = wr.ID
		m.Request = wr.Request

		wr.ID = 0
		wr.Request = nil

		select {
		case workersCh <- struct{}{}:
		default:
			select {
			case workersCh <- struct{}{}:
			case <-stopChan:
				return
			}
		}
		go serveRequest(s, responsesChan, stopChan, m, workersCh)
	}
}

func serveRequest(s *Server, responsesChan chan<- *serverMessage, stopChan <-chan struct{}, m *serverMessage, workersCh <-chan struct{}) {
	request := m.Request
	m.Request = nil
	skipResponse := m.ID == 0

	if skipResponse {
		m.Response = nil
		m.Error = ""
		serverMessagePool.Put(m)
	}

	response, err := callHandlerWithRecover(s.LogError, s.Handler, s.Addr, request)

	if !skipResponse {
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
	}

	<-workersCh
}

func callHandlerWithRecover(logErrorFunc LoggerFunc, handler HandlerFunc, serverAddr string, request interface{}) (response interface{}, errStr string) {
	defer func() {
		if x := recover(); x != nil {
			stackTrace := make([]byte, 1<<20)
			n := runtime.Stack(stackTrace, false)
			errStr = fmt.Sprintf("Panic occured: %v\nStack trace: %s", x, stackTrace[:n])
			logErrorFunc("ztcp.Server: %s", errStr)
		}
	}()
	response = handler(request)
	return
}

func serverWriter(s *Server, w io.Writer, responsesChan <-chan *serverMessage, stopChan <-chan struct{}, done chan<- struct{}) {
	defer func() { close(done) }()

	e := newMessageEncoder(w, s.SendBufferSize)
	defer e.Close()

	var flushChan <-chan time.Time
	t := time.NewTimer(s.FlushDelay)
	var wr wireResponse
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
				if err := e.Flush(); err != nil {
					if !isServerStop(stopChan) {
						s.LogError("ztcp.Server: Cannot flush responses to underlying stream: [%s]", err)
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

		wr.ID = m.ID
		wr.Response = m.Response
		wr.Error = m.Error

		m.Response = nil
		m.Error = ""
		serverMessagePool.Put(m)

		if err := e.Encode(wr); err != nil {
			s.LogError("ztcp.Server: Cannot send response to wire: [%s]", err)
			return
		}
		wr.Response = nil
		wr.Error = ""

	}
}
