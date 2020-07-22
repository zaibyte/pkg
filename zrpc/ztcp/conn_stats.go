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
	"sync"
	"time"
)

// ConnStats provides connection statistics. Applied to both gorpc.Client
// and gorpc.Server.
//
// Use stats returned from ConnStats.Snapshot() on live Client and / or Server,
// since the original stats can be updated by concurrently running goroutines.
type ConnStats struct {
	// The number of rpc calls performed.
	RPCCalls uint64

	// The total aggregate time for all rpc calls in milliseconds.
	//
	// This time can be used for calculating the average response time
	// per RPC:
	//     avgRPCTtime = RPCTime / RPCCalls
	RPCTime uint64

	// The number of bytes written to the underlying connections.
	BytesWritten uint64

	// The number of bytes read from the underlying connections.
	BytesRead uint64

	// The number of Read() calls.
	ReadCalls uint64

	// The number of Read() errors.
	ReadErrors uint64

	// The number of Write() calls.
	WriteCalls uint64

	// The number of Write() errors.
	WriteErrors uint64

	// The number of Dial() calls.
	DialCalls uint64

	// The number of Dial() errors.
	DialErrors uint64

	// The number of Accept() calls.
	AcceptCalls uint64

	// The number of Accept() errors.
	AcceptErrors uint64

	// lock is for 386 builds. See https://github.com/valyala/gorpc/issues/5 .
	lock sync.Mutex
}

// AvgRPCTime returns the average RPC execution time.
//
// Use stats returned from ConnStats.Snapshot() on live Client and / or Server,
// since the original stats can be updated by concurrently running goroutines.
func (cs *ConnStats) AvgRPCTime() time.Duration {
	return time.Duration(float64(cs.RPCTime)/float64(cs.RPCCalls)) * time.Millisecond
}

// AvgRPCBytes returns the average bytes sent / received per RPC.
//
// Use stats returned from ConnStats.Snapshot() on live Client and / or Server,
// since the original stats can be updated by concurrently running goroutines.
func (cs *ConnStats) AvgRPCBytes() (send float64, recv float64) {
	return float64(cs.BytesWritten) / float64(cs.RPCCalls), float64(cs.BytesRead) / float64(cs.RPCCalls)
}

// AvgRPCCalls returns the average number of write() / read() syscalls per PRC.
//
// Use stats returned from ConnStats.Snapshot() on live Client and / or Server,
// since the original stats can be updated by concurrently running goroutines.
func (cs *ConnStats) AvgRPCCalls() (write float64, read float64) {
	return float64(cs.WriteCalls) / float64(cs.RPCCalls), float64(cs.ReadCalls) / float64(cs.RPCCalls)
}
