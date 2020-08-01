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
// Package xdigest provides hash functions on byte sequences by wrapping xxhash.
// These hash functions are intended to be used to implement zai object digest
// that need to map byte sequences to a uniform distribution on unsigned 32-bit integers.
//
// For Zai, the needs of object digest:
// 1. Fast.
// candidates: xxh3, xxhash, hashes based on AES, crc32
// 2. Low collisions with 32bits sum.
// candidates: xxh3_low, xxhash_low (hashes based on AES can't pass smhasher sparse test, crc32 can't pass smhasher avalanche test)
// 3. Stable
// candidates: xxh3 (may not change after v0.7.4), xxhash
// 4. Could satisfy hash.Hash32 interface.
// candidates: xxhash (xxh3 only has one-shot hash API)
package xdigest

import (
	"github.com/cespare/xxhash/v2"
	"github.com/zaibyte/pkg/xstrconv"
)

type digest struct {
	xxh *xxhash.Digest
}

// New creates a xdigest.
func New() *digest {
	return &digest{xxhash.New()}
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *digest) Write(b []byte) (n int, err error) {
	return d.xxh.Write(b)
}

// WriteString (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *digest) WriteString(s string) (n int, err error) {
	return d.Write(xstrconv.ToBytes(s))
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *digest) Sum(b []byte) []byte {
	s := d.Sum32()
	return append(
		b,
		byte(s>>24),
		byte(s>>16),
		byte(s>>8),
		byte(s),
	)
}

// Sum32 returns the current hash.
func (d *digest) Sum32() uint32 {
	return uint32(d.xxh.Sum64())
}

// Reset resets the Hash to its initial state.
func (d *digest) Reset() {
	d.xxh.Reset()
}

// Size returns the number of bytes Sum will return.
func (d *digest) Size() int {
	return 4
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (d *digest) BlockSize() int {
	return 32
}

// Sum32 computes the 32-bit xxHash_low32bit digest of b.
func Sum32(b []byte) uint32 {
	return uint32(xxhash.Sum64(b))
}

// Sum32String computes the 32-bit xxHash_low32bit digest of b.
func Sum32String(s string) uint32 {
	return Sum32(xstrconv.ToBytes(s))
}
