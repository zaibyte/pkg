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

package xtcp

import (
	"math/rand"
	"testing"

	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/xrpc"
)

func BenchmarkZtcpByteSlice(b *testing.B) {
	addr := getRandomAddr()

	r := NewRouter()
	r.AddFunc(1, bytesHandlerFunc)

	s := NewServer(addr, r, nil)
	if err := s.Start(); err != nil {
		b.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	r.AddToClient(c)
	c.Start()
	defer c.Stop()

	req := make([]byte, 4096)
	rand.Read(req)
	xbuf := xrpc.GetBytes()
	xbuf.Write(req)
	b.SetParallelism(1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			resp := xrpc.GetBytes()
			err := c.Call(uid.MakeReqID(), 1, nil, resp, xbuf)
			if err != nil {
				b.Fatalf("Unexpected error: %s", err)
			}

			resp.Free()
		}
	})
}
