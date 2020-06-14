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
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseReqID(t *testing.T) {

	startTicker()
	defer StopTicker()

	reqids := new(sync.Map)
	// Because it's fast, second ts won't change.
	expTime := Ts2Time(atomic.LoadUint32(&ticker.ts))

	wg := new(sync.WaitGroup)
	n := runtime.NumCPU()
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(seed int) {
			defer wg.Done()

			boxID := uint32(seed + 1)
			reqid := MakeReqID(boxID)
			reqids.Store(boxID, reqid)

		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		boxID := uint32(i + 1)
		v, ok := reqids.Load(boxID)
		assert.True(t, ok)
		reqID := v.(string)
		actBoxID, actTime, err := ParseReqID(reqID)
		assert.Nil(t, err)
		assert.Equal(t, boxID, actBoxID)
		assert.Equal(t, expTime, actTime)
	}
}

func BenchmarkMakeReqID(b *testing.B) {

	startTicker()
	defer StopTicker()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = MakeReqID(1)
	}
}

func BenchmarkMakeReqID_Parallel(b *testing.B) {

	startTicker()
	defer StopTicker()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MakeReqID(1)
		}
	})
}

func BenchmarkParseReqID(b *testing.B) {

	startTicker()
	defer StopTicker()

	b.ResetTimer()

	s := MakeReqID(1)

	for i := 0; i < b.N; i++ {
		_, _, _ = ParseReqID(s)
	}
}

func BenchmarkParseReqID_Parallel(b *testing.B) {

	startTicker()
	defer StopTicker()

	b.ResetTimer()

	s := MakeReqID(1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = ParseReqID(s)
		}
	})
}
