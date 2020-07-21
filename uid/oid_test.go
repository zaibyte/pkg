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

func TestOID(t *testing.T) {

	oidStrs := new(sync.Map)

	// Because it's fast, second ts won't change.
	expTime := atomic.LoadUint32(&ticker.ts)

	wg := new(sync.WaitGroup)
	n := runtime.NumCPU()
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(seed int) {
			defer wg.Done()

			boxID := uint32(seed + 1)
			extID := uint32(seed + 2)
			digest := uint32(seed + 3)
			size := uint32(seed + 4)
			otype := uint8(seed & 7)

			_, oid := MakeOID(boxID, extID, digest, size, otype)
			oidStrs.Store(seed, oid)

		}(i)
	}
	wg.Wait()

	wg2 := new(sync.WaitGroup)
	wg2.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg2.Done()

			v, ok := oidStrs.Load(i)
			assert.True(t, ok)

			oid := v.(string)
			boxID, extID, ts, digest, size, otype, err := ParseOID(oid)
			assert.Nil(t, err)

			expboxID := uint32(i + 1)
			expextID := uint32(i + 2)
			expdigest := uint32(i + 3)
			expsize := uint32(i + 4)
			expotype := uint8(i & 7)

			assert.Equal(t, expboxID, boxID)
			assert.Equal(t, expextID, extID)
			assert.Equal(t, expTime, ts)
			assert.Equal(t, expdigest, digest)
			assert.Equal(t, expsize, size)
			assert.Equal(t, expotype, otype)
		}(i)
	}
	wg2.Wait()
}

func BenchmarkMakeOID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_, _ = MakeOID(1, 1, 1, 1, 1)
	}
}

func BenchmarkMakeOID_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = MakeOID(1, 1, 1, 1, 1)
		}
	})
}

func BenchmarkParseOID(b *testing.B) {

	_, oid := MakeOID(1, 2, 3, 4, 1)

	for i := 0; i < b.N; i++ {
		_, _, _, _, _, _, _ = ParseOID(oid)
	}
}

func BenchmarkParseOID_Parallel(b *testing.B) {

	_, oid := MakeOID(1, 2, 3, 4, 1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _, _, _, _, _ = ParseOID(oid)
		}
	})
}
