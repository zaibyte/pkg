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
	"io"
	"net"
	"sync"
	"time"

	"github.com/zaibyte/pkg/xerrors"

	"github.com/zaibyte/pkg/xrpc"

	"github.com/zaibyte/pkg/xlog"
)

const (
	objPutMethod uint8 = 1
	objGetMethod uint8 = 2
	objDelMethod uint8 = 3
)

const (
	// DefaultPendingMessages is the default number of pending messages
	// handled by Client and Server.
	DefaultPendingMessages = 32 * 1024

	// DefaultFlushDelay is the default delay between message flushes
	// on Client and Server.
	DefaultFlushDelay = time.Microsecond * 100
)

var timerPool sync.Pool

func acquireTimer(timeout time.Duration) *time.Timer {
	tv := timerPool.Get()
	if tv == nil {
		return time.NewTimer(timeout)
	}

	t := tv.(*time.Timer)
	if t.Reset(timeout) {
		xlog.Panic("bug: active timer trapped into acquireTimer()")
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

func readBytes(r net.Conn, buf []byte, perRead int, encrypted bool, hash hash.Hash32, digest uint32) (err error) {

	n := len(buf)
	received := 0
	var recvBuf []byte
	if n < perRead {
		recvBuf = buf[:n]
	} else {
		recvBuf = buf[:perRead]
	}
	toRead := n
	deadline := time.Now()
	for toRead > 0 {
		deadline = deadline.Add(readDuration)
		if err = r.SetReadDeadline(deadline); err != nil {
			return fmt.Errorf("failed to set read deadline: %s, %s", r.RemoteAddr().String(), err.Error())
		}

		if _, err = io.ReadFull(r, recvBuf); err != nil {
			return fmt.Errorf("failed to read: %s, %s", r.RemoteAddr().String(), err.Error())
		}
		if !encrypted {
			hash.Write(recvBuf)
		}
		toRead -= len(recvBuf)
		received += len(recvBuf)
		if toRead < perRead {
			recvBuf = buf[received : received+toRead]
		} else {
			recvBuf = buf[received : received+perRead]
		}
	}
	if received != n {
		return fmt.Errorf("unexpected received size: %d, but want %d", received, n)
	}
	actDigest := hash.Sum32()
	if !encrypted && actDigest != digest {
		return xerrors.WithMessage(xrpc.ErrChecksumMismatch, fmt.Sprintf("exp: %d, but: %d", digest, actDigest))
	}
	return
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
