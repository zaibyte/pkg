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

	"github.com/stretchr/testify/assert"
)

func TestOID(t *testing.T) {

	oids := make([]*OID, 10)
	for i := range oids {
		oids[i] = &OID{
			BoxID:    uint32(i + 1),
			ExtentID: uint32(i + 2),
			Digest:   uint64(i + 3),
			Size:     uint32(i + 4),
		}
	}

	for i, o := range oids {
		o := o
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			v := o.Marshal()
			err := o.Unmarshal(v)
			if err != nil {
				t.Fatal(err)
			}

			o2 := new(OID)
			err = o2.Unmarshal(v)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, o, o2)
		})
	}
}

func BenchmarkOID_Marshal(b *testing.B) {

	oid := &OID{
		BoxID:    1,
		ExtentID: 1,
		Digest:   1,
		Size:     1,
	}

	for i := 0; i < b.N; i++ {
		_ = oid.Marshal()
	}
}

func BenchmarkOID_Marshal_Parallel(b *testing.B) {

	oid := &OID{
		BoxID:    1,
		ExtentID: 1,
		Digest:   1,
		Size:     1,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = oid.Marshal()
		}
	})
}

func BenchmarkOID_Unmarshal(b *testing.B) {

	oid := &OID{
		BoxID:    1,
		ExtentID: 1,
		Digest:   1,
		Size:     1,
	}
	s := oid.Marshal()

	for i := 0; i < b.N; i++ {
		_ = oid.Unmarshal(s)
	}
}

func BenchmarkOID_Unmarshal_Parallel(b *testing.B) {

	oid := &OID{
		BoxID:    1,
		ExtentID: 1,
		Digest:   1,
		Size:     1,
	}
	s := oid.Marshal()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = oid.Unmarshal(s)
		}
	})
}
