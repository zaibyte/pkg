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

// +build !386

package ztcp

import (
	"sync/atomic"
)

// Snapshot returns connection statistics' snapshot.
//
// Use stats returned from ConnStats.Snapshot() on live Client and / or Server,
// since the original stats can be updated by concurrently running goroutines.
func (cs *ConnStats) Snapshot() *ConnStats {
	return &ConnStats{
		RPCCalls:     atomic.LoadUint64(&cs.RPCCalls),
		RPCTime:      atomic.LoadUint64(&cs.RPCTime),
		BytesWritten: atomic.LoadUint64(&cs.BytesWritten),
		BytesRead:    atomic.LoadUint64(&cs.BytesRead),
		ReadCalls:    atomic.LoadUint64(&cs.ReadCalls),
		ReadErrors:   atomic.LoadUint64(&cs.ReadErrors),
		WriteCalls:   atomic.LoadUint64(&cs.WriteCalls),
		WriteErrors:  atomic.LoadUint64(&cs.WriteErrors),
		DialCalls:    atomic.LoadUint64(&cs.DialCalls),
		DialErrors:   atomic.LoadUint64(&cs.DialErrors),
		AcceptCalls:  atomic.LoadUint64(&cs.AcceptCalls),
		AcceptErrors: atomic.LoadUint64(&cs.AcceptErrors),
	}
}

// Reset resets all the stats counters.
func (cs *ConnStats) Reset() {
	atomic.StoreUint64(&cs.RPCCalls, 0)
	atomic.StoreUint64(&cs.RPCTime, 0)
	atomic.StoreUint64(&cs.BytesWritten, 0)
	atomic.StoreUint64(&cs.BytesRead, 0)
	atomic.StoreUint64(&cs.WriteCalls, 0)
	atomic.StoreUint64(&cs.WriteErrors, 0)
	atomic.StoreUint64(&cs.ReadCalls, 0)
	atomic.StoreUint64(&cs.ReadErrors, 0)
	atomic.StoreUint64(&cs.DialCalls, 0)
	atomic.StoreUint64(&cs.DialErrors, 0)
	atomic.StoreUint64(&cs.AcceptCalls, 0)
	atomic.StoreUint64(&cs.AcceptErrors, 0)
}

func (cs *ConnStats) incRPCCalls() {
	atomic.AddUint64(&cs.RPCCalls, 1)
}

func (cs *ConnStats) incRPCTime(dt uint64) {
	atomic.AddUint64(&cs.RPCTime, dt)
}

func (cs *ConnStats) addBytesWritten(n uint64) {
	atomic.AddUint64(&cs.BytesWritten, n)
}

func (cs *ConnStats) addBytesRead(n uint64) {
	atomic.AddUint64(&cs.BytesRead, n)
}

func (cs *ConnStats) incReadCalls() {
	atomic.AddUint64(&cs.ReadCalls, 1)
}

func (cs *ConnStats) incReadErrors() {
	atomic.AddUint64(&cs.ReadErrors, 1)
}

func (cs *ConnStats) incWriteCalls() {
	atomic.AddUint64(&cs.WriteCalls, 1)
}

func (cs *ConnStats) incWriteErrors() {
	atomic.AddUint64(&cs.WriteErrors, 1)
}

func (cs *ConnStats) incDialCalls() {
	atomic.AddUint64(&cs.DialCalls, 1)
}

func (cs *ConnStats) incDialErrors() {
	atomic.AddUint64(&cs.DialErrors, 1)
}

func (cs *ConnStats) incAcceptCalls() {
	atomic.AddUint64(&cs.AcceptCalls, 1)
}

func (cs *ConnStats) incAcceptErrors() {
	atomic.AddUint64(&cs.AcceptErrors, 1)
}
