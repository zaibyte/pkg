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
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	// DefaultConcurrency is the default number of concurrent rpc calls
	// the server can process.
	DefaultConcurrency = 8 * 1024

	// DefaultRequestTimeout is the default timeout for client request.
	DefaultRequestTimeout = 5 * time.Second

	// DefaultPendingMessages is the default number of pending messages
	// handled by Client and Server.
	DefaultPendingMessages = 32 * 1024

	// DefaultFlushDelay is the default delay between message flushes
	// on Client and Server.
	DefaultFlushDelay = -1

	// DefaultBufferSize is the default size for Client and Server buffers.
	DefaultBufferSize = 64 * 1024
)

// OnConnectFunc is a callback, which may be called by both Client and Server
// on every connection creation if assigned
// to Client.OnConnect / Server.OnConnect.
//
// remoteAddr is the address of the remote end for the established
// connection rwc.
//
// The callback must return either rwc itself or a rwc wrapper.
// The returned connection wrapper MUST send all the data to the underlying
// rwc on every Write() call, otherwise the connection will hang forever.
//
// The callback may be used for authentication/authorization and/or custom
// transport wrapping.
type OnConnectFunc func(remoteAddr string, rwc io.ReadWriteCloser) (net.Conn, error)

// LoggerFunc is an error logging function to pass to gorpc.SetErrorLogger().
type LoggerFunc func(format string, args ...interface{})

var errorLogger = LoggerFunc(log.Printf)

// SetErrorLogger sets the given error logger to use in gorpc.
//
// By default log.Printf is used for error logging.
func SetErrorLogger(f LoggerFunc) {
	errorLogger = f
}

// NilErrorLogger discards all error messages.
//
// Pass NilErrorLogger to SetErrorLogger() in order to suppress error log generated
// by gorpc.
func NilErrorLogger(format string, args ...interface{}) {}

func logPanic(format string, args ...interface{}) {
	errorLogger(format, args...)
	s := fmt.Sprintf(format, args...)
	panic(s)
}

var timerPool sync.Pool

func acquireTimer(timeout time.Duration) *time.Timer {
	tv := timerPool.Get()
	if tv == nil {
		return time.NewTimer(timeout)
	}

	t := tv.(*time.Timer)
	if t.Reset(timeout) {
		panic("BUG: Active timer trapped into acquireTimer()")
	}
	return t
}

func releaseTimer(t *time.Timer) {
	if !t.Stop() {
		// Collect possibly added time from the channel
		// if timer has been stopped and nobody collected its' value.
		select {
		case <-t.C:
		default:
		}
	}

	timerPool.Put(t)
}

var closedFlushChan = make(chan time.Time)

func init() {
	close(closedFlushChan)
}

func getFlushChan(t *time.Timer, flushDelay time.Duration) <-chan time.Time {
	if flushDelay <= 0 {
		return closedFlushChan
	}

	if !t.Stop() {
		// Exhaust expired timer's chan.
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(flushDelay)
	return t.C
}
