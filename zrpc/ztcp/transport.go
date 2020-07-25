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
	"crypto/tls"
	"errors"
	"net"
	"time"
)

var (
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrBadMessage is the error returned to indicate the incoming message is corrupted.
	ErrBadMessage     = errors.New("invalid message")
	ErrPoisonReceived = errors.New("poison received")
	ErrInvalidMethod  = errors.New("invalid method")
)

var (
	magicNumber         = [2]byte{0x7A, 0x61}
	poisonNumber        = [2]byte{0x0, 0x0}
	payloadBufferSize   = 4 * 1024 * 1024
	dialTimeout         = 2 * time.Second
	tlsHandshakeTimeout = 3 * time.Second
	magicNumberDuration = 1 * time.Second
	headerDuration      = 2 * time.Second
	readDuration        = 5 * time.Second
	writeDuration       = 5 * time.Second
	keepAlivePeriod     = 10 * time.Second
)

var (
	dialer = &net.Dialer{
		Timeout: dialTimeout,
	}
)

// DialFunc is a function intended for setting to Client.Dial.
//
// It is expected that the returned conn immediately
// sends all the data passed via Write() to the server.
// Otherwise ztcp may hang.
// The conn implementation must call Flush() on underlying buffered
// streams before returning from Write().
type DialFunc func(addr string) (conn net.Conn, err error)

// Listener is an interface for custom listeners intended for the Server.
type Listener interface {
	// Init is called on server start.
	//
	// addr contains the address set at Server.Addr.
	Init(addr string) error

	// Accept must return incoming connections from clients.
	//
	// It is expected that the returned conn immediately
	// sends all the data passed via Write() to the client.
	// Otherwise ztcp may hang.
	// The conn implementation must call Flush() on underlying buffered
	// streams before returning from Write().
	Accept() (conn net.Conn, err error)

	// Close closes the listener.
	// All pending calls to Accept() must immediately return errors after
	// Close is called.
	// All subsequent calls to Accept() must immediately return error.
	Close() error

	// Addr returns the listener's network address.
	ListenAddr() net.Addr
}

func defaultDial(addr string) (conn net.Conn, err error) {
	return dialer.Dial("tcp", addr)
}

type defaultListener struct {
	L      net.Listener
	tlsCfg *tls.Config
}

func (ln *defaultListener) Init(addr string) (err error) {
	ln.L, err = net.Listen("tcp", addr)
	return
}

func (ln *defaultListener) ListenAddr() net.Addr {
	if ln.L != nil {
		return ln.L.Addr()
	}
	return nil
}

func (ln *defaultListener) Accept() (conn net.Conn, err error) {
	c, err := ln.L.Accept()
	if err != nil {
		return nil, err
	}
	tcpConn := c.(*net.TCPConn)
	if err = setTCPConn(tcpConn); err != nil {
		c.Close()
		return nil, err
	}
	if ln.tlsCfg != nil {
		c = tls.Server(c, ln.tlsCfg)
		tt := time.Now().Add(tlsHandshakeTimeout)
		if err := c.SetDeadline(tt); err != nil {
			return nil, err
		}
		if err := c.(*tls.Conn).Handshake(); err != nil { // Do handshake manually before ztcp build connection.
			return nil, err
		}
		if err := c.SetDeadline(time.Time{}); err != nil { // Reset deadline because conn will be reused.
			return nil, err
		}
	}
	return c, nil
}

func (ln *defaultListener) Close() error {
	return ln.L.Close()
}

// NewTCPServer creates a server listening for TLS (if has) or TCP connections
// on the given addr and processing incoming requests
// with the given HandlerFunc.
//
// The returned server must be started after optional settings' adjustment.
//
// The corresponding client must be created with NewClient().
func NewServer(addr string, handler HandlerFunc, cfg *tls.Config) *Server {
	if cfg == nil {
		return NewTCPServer(addr, handler)
	}
	return NewTLSServer(addr, handler, cfg)
}

// NewTCPServer creates a server listening for TCP connections
// on the given addr and processing incoming requests
// with the given HandlerFunc.
//
// The returned server must be started after optional settings' adjustment.
//
// The corresponding client must be created with NewTCPClient().
func NewTCPServer(addr string, handler HandlerFunc) *Server {
	return &Server{
		Addr:     addr,
		Handler:  handler,
		Listener: &defaultListener{},
	}
}

// NewTLSServer creates a server listening for TLS (aka SSL) connections
// on the given addr and processing incoming requests
// with the given HandlerFunc.
// cfg must contain TLS settings for the server.
//
// The returned server must be started after optional settings' adjustment.
//
// The corresponding client must be created with NewTLSClient().
func NewTLSServer(addr string, handler HandlerFunc, cfg *tls.Config) *Server {
	return &Server{
		Addr:    addr,
		Handler: handler,
		Listener: &defaultListener{
			tlsCfg: cfg,
		},
	}
}

// NewClient creates a client connecting over TLS (if has) or TCP
// to the server listening to the given addr.
//
// The returned client must be started after optional settings' adjustment.
//
// The corresponding server must be created with NewServer().
func NewClient(addr string, cfg *tls.Config) *Client {
	if cfg == nil {
		return NewTCPClient(addr)
	}
	return NewTLSClient(addr, cfg)
}

// NewTCPClient creates a client connecting over TCP to the server
// listening to the given addr.
//
// The returned client must be started after optional settings' adjustment.
//
// The corresponding server must be created with NewTCPServer().
func NewTCPClient(addr string) *Client {
	return &Client{
		Addr: addr,
		Dial: func(addr string) (conn net.Conn, err error) {
			return getConnection(addr, nil)
		},
	}
}

// NewTLSClient creates a client connecting over TLS (aka SSL) to the server
// listening to the given addr using the given TLS config.
//
// The returned client must be started after optional settings' adjustment.
//
// The corresponding server must be created with NewTLSServer().
func NewTLSClient(addr string, cfg *tls.Config) *Client {
	return &Client{
		Addr: addr,
		Dial: func(addr string) (conn net.Conn, err error) {
			return getConnection(addr, cfg)
		},
	}
}

func setTCPConn(conn *net.TCPConn) error {
	if err := conn.SetLinger(0); err != nil {
		return err
	}
	if err := conn.SetKeepAlive(true); err != nil {
		return err
	}
	return conn.SetKeepAlivePeriod(keepAlivePeriod)
}

func getConnection(target string, tlsConfig *tls.Config) (net.Conn, error) {
	conn, err := dialer.Dial("tcp", target)
	if err != nil {
		return nil, err
	}
	tcpconn, ok := conn.(*net.TCPConn)
	if ok {
		if err := setTCPConn(tcpconn); err != nil {
			return nil, err
		}
	}

	if tlsConfig != nil {
		conn = tls.Client(conn, tlsConfig)
		tt := time.Now().Add(tlsHandshakeTimeout)
		if err := conn.SetDeadline(tt); err != nil {
			return nil, err
		}
		if err := conn.(*tls.Conn).Handshake(); err != nil {
			return nil, err
		}
		if err := conn.SetDeadline(time.Time{}); err != nil {
			return nil, err
		}
	}
	return conn, nil
}
