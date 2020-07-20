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

package uid

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/templexxx/tsc"

	"github.com/stretchr/testify/assert"
)

func TestParseReqID(t *testing.T) {

	reqids := new(sync.Map)
	// Because it's fast, second ts won't change.
	expTime := time.Unix(0, tsc.UnixNano())

	wg := new(sync.WaitGroup)
	n := runtime.NumCPU()
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(seed int) {
			defer wg.Done()

			boxID := uint32(seed + 1)
			reqid := MakeReqID()
			reqids.Store(boxID, reqid)

		}(i)
	}
	wg.Wait()

	wg2 := new(sync.WaitGroup)
	wg2.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg2.Done()

			boxID := uint32(i + 1)
			v, ok := reqids.Load(boxID)
			assert.True(t, ok)
			reqID := v.(uint64)
			actTime := ParseReqID(reqID)
			assert.Equal(t, expTime.Unix(), actTime.Unix())
		}(i)
	}
	wg2.Wait()
}

func BenchmarkMakeReqID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeReqID()
	}
}

func BenchmarkMakeReqID_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MakeReqID()
		}
	})
}

func BenchmarkParseReqID(b *testing.B) {

	s := MakeReqID()

	for i := 0; i < b.N; i++ {
		_ = ParseReqID(s)
	}
}

func BenchmarkParseReqID_Parallel(b *testing.B) {

	s := MakeReqID()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ParseReqID(s)
		}
	})
}
