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
	"math/rand"
	"testing"
	"time"

	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/xdigest"
)

func BenchmarkClient_Put(b *testing.B) {

	rand.Seed(time.Now().UnixNano())

	addr := getRandomAddr()

	s := NewServer(addr, nil, testPutFunc, testGetFunc, testDeleteFunc)
	if err := s.Start(); err != nil {
		b.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	req := make([]byte, 4096)
	rand.Read(req)
	digest := xdigest.Sum32(req)
	_, oid := uid.MakeOID(1, 1, digest, 4096, uid.NormalObj)

	b.SetParallelism(256)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			err := c.Put(uid.MakeReqID(), oid, req, 0)
			if err != nil {
				b.Fatalf("Unexpected error: %s", err)
			}

		}
	})
}

func BenchmarkClient_Delete(b *testing.B) {

	rand.Seed(time.Now().UnixNano())

	addr := getRandomAddr()

	s := NewServer(addr, nil, testPutFunc, testGetFunc, testDeleteFunc)
	if err := s.Start(); err != nil {
		b.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	req := make([]byte, 4096)
	rand.Read(req)
	digest := xdigest.Sum32(req)
	_, oid := uid.MakeOID(1, 1, digest, 4096, uid.NormalObj)

	b.SetParallelism(256)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			err := c.Delete(uid.MakeReqID(), oid, 0)
			if err != nil {
				b.Fatalf("Unexpected error: %s", err)
			}

		}
	})
}
