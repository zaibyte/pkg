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

package zdigest

import (
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"hash/maphash"
	"math/rand"
	"testing"

	"github.com/mengzhuo/nabhash"
	"github.com/templexxx/xorsimd"
	"github.com/zeebo/xxh3"
)

func BenchmarkCRC(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	table := crc32.MakeTable(crc32.Castagnoli)
	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		_ = crc32.Checksum(p, table)
	}
}

func BenchmarkAlder(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		_ = adler32.Checksum(p)
	}
}

func TestDigest_Sum32(t *testing.T) {
	for i := 0; i < 256; i++ {
		fmt.Println(Sum32([]byte{uint8(i)}))
	}
}
func BenchmarkZDigest(b *testing.B) {

	for i := 8; i <= 65536; i <<= 1 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			buf := make([]byte, i)
			rand.Read(buf)
			b.SetBytes(int64(i))
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = Sum32(buf)
			}
			return
		})
	}
}

func BenchmarkXXHALL(b *testing.B) {

	for i := 8; i <= 65536; i <<= 1 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			buf := make([]byte, i)
			rand.Read(buf)
			b.SetBytes(int64(i))
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = xxh3.Hash(buf)
			}
			return
		})
	}
}

func BenchmarkXXH3(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		_ = xxh3.Hash(p)
	}
}

func BenchmarkMapHash(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	mh := &maphash.Hash{}

	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		mh.Write(p)
		_ = mh.Sum64()
	}
}

func BenchmarkNAB(b *testing.B) {

	p := make([]byte, 1024*1024)
	s := make([]byte, 16)
	rand.Read(p)

	mh := nabhash.New()

	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		mh.Write(p)
		sum := mh.Sum(s[:0])
		xorsimd.Bytes8(sum, sum[:8], sum[8:])
		s64 := binary.LittleEndian.Uint64(sum[:8])
		_ = s64>>4 ^ s64
		mh.Reset()
	}
}

func BenchmarkDSum32(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	b.SetBytes(1024 * 1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Sum32(p)
	}
}

func TestDSum32(t *testing.T) {

	p := make([]byte, 1024*1024)
	rand.Read(p)

	ds := Sum32(p)
	d := New()
	d.Write(p)
	dds := d.Sum32()
	fmt.Println(ds, dds)
}

func BenchmarkDigest_Sum32(b *testing.B) {

	p := make([]byte, 1024*1024)
	rand.Read(p)
	d := New()

	b.ResetTimer()
	b.SetBytes(1024 * 1024)

	for i := 0; i < b.N; i++ {
		d.Write(p)
		_ = d.Sum32()
		d.Reset()
	}
}
