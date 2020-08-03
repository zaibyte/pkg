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

//func BenchmarkZtcpByteSlice(b *testing.B) {
//	addr := getRandomAddr()
//
//	r := NewRouter()
//	r.AddFunc(1, bytesHandlerFunc)
//
//	s := NewServer(addr, r, nil)
//	if err := s.Start(); err != nil {
//		b.Fatalf("cannot start server: %s", err)
//	}
//	defer s.Stop()
//
//	c := NewClient(addr, nil)
//	r.AddToClient(c)
//	c.Start()
//	defer c.Stop()
//
//	req := make([]byte, 4096)
//	rand.Read(req)
//	xbuf := xrpc.GetBytes()
//	defer xbuf.Free()
//	xbuf.Write(req)
//	b.SetParallelism(250)
//	b.ResetTimer()
//	b.RunParallel(func(pb *testing.PB) {
//		for i := 0; pb.Next(); i++ {
//			resp := xrpc.GetBytes()
//			err := c.call(uid.MakeReqID(), 1, nil, resp, xbuf)
//			if err != nil {
//				b.Fatalf("Unexpected error: %s", err)
//			}
//
//			if !bytes.Equal(resp.Bytes(), req) {
//				b.Fatal("Unexpected response")
//			}
//
//			resp.Free()
//		}
//	})
//}
