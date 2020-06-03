/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xlog_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/zaibyte/pkg/uid"

	"github.com/zaibyte/pkg/xlog/xlogtest"
)

func BenchmarkAccessLogger_Write(b *testing.B) {
	_, al := xlogtest.New("xlog-acc")
	defer func() {
		xlogtest.Clean()
	}()
	req, err := http.NewRequest(http.MethodGet, "http://1.com", nil)
	if err != nil {
		b.Fatal(err)
	}
	start := time.Now()
	reqID := uid.MakeReqIDWithTime(0, start)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.Write("-", req, start, reqID, 1, 200)
	}
}

func BenchmarkAccessLogger_Write_Parallel(b *testing.B) {

	_, al := xlogtest.New("xlog-acc")
	defer func() {
		xlogtest.Clean()
	}()
	req, err := http.NewRequest(http.MethodGet, "http://1.com", nil)
	if err != nil {
		b.Fatal(err)
	}
	start := time.Now()
	reqID := uid.MakeReqIDWithTime(0, start)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			al.Write("-", req, start, reqID, 1, 200)
		}
	})
}
