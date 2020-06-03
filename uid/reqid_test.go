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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseReqID(t *testing.T) {

	times := make([]time.Time, 10)
	for i := range times {
		times[i] = time.Unix(int64(i), 0)
	}

	for i, ti := range times {
		ti := ti
		expBoxID := uint32(i + 1)
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			reqID := MakeReqIDWithTime(expBoxID, ti)
			boxID, pid, pt, err := ParseReqID(reqID)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, expBoxID, boxID)
			assert.Equal(t, _pid, pid)
			assert.Equal(t, ti, pt)
		})
	}
}

func BenchmarkMakeReqID(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = MakeReqID(1)
	}
}

func BenchmarkMakeReqID_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MakeReqID(1)
		}
	})
}

func BenchmarkParseReqID(b *testing.B) {

	s := MakeReqID(1)

	for i := 0; i < b.N; i++ {
		_, _, _, _ = ParseReqID(s)
	}
}

func BenchmarkParseReqID_Parallel(b *testing.B) {

	s := MakeReqID(1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _, _ = ParseReqID(s)
		}
	})
}