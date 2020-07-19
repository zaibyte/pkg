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
// BSD 3-Clause License
//
// Copyright (c) 2019, Meng Zhuo
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// This file contains code derived from NABHash.
// The main logic is copied from NABHash,
// replacing 128bit digest with 32bit unsigned integer.

// Package zdigest provides hash functions on byte sequences.
// These hash functions are intended to be used to implement zai object digest
// that need to map byte sequences to a uniform distribution on unsigned 32-bit integers.
//
// For Zai, the needs of object digest:
// 1. Extremely fast for big data (>4KB).
// 2. Low collisions.
//
// Unlike hash.Hash, the zdigest usage:
// 1.
// d := New()
// d.Write()
// d.Sum32()
// d.Reset()
// d.Write()
// d.Sum32()
// ...
// 2.
// Sum32()
package zdigest

const (
	// Size of zdigest in bytes.
	Size = 4

	// BlockSize of zdigest in bytes.
	BlockSize = 64

	blockSizeMask = BlockSize - 1
)

var zeroData = make([]byte, BlockSize)

var initState = state{
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
}

var initStateBytes = []byte{
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
	0x5A, 0x82, 0x79, 0x99, 0x6E, 0xD9, 0xEB, 0xA1,
	0x8F, 0x1B, 0xBC, 0xDC, 0xCA, 0x62, 0xC1, 0xD6,
}

var block = blockGeneric
var final = finalGeneric
var sum32 = sum32Generic

type state [BlockSize]byte

type Digest struct {
	h      state
	buf    state
	length uint64
	remain int
}

// New return a new Digest computing the zdigest checksum.
func New() *Digest {
	d := &Digest{}
	d.Reset()
	return d
}

// Write adds b to the sequence of bytes hashed by h.
// It always writes all of b and never fails; the count and error result are for implementing io.Writer.
func (d *Digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	if d.remain > 0 {
		n := copy(d.buf[d.remain:], p)
		d.remain += n
		if d.remain == BlockSize {
			block(&d.h, d.buf[:])
			d.remain -= BlockSize
		}
		p = p[n:]
	}

	if len(p) >= BlockSize {
		n := len(p) &^ blockSizeMask
		block(&d.h, p[:n])
		p = p[n:]
	}

	if len(p) > 0 {
		d.remain = copy(d.buf[:], p)
	}
	d.length += uint64(nn)
	return
}

// Sum32 returns hash's current 32-bit value.
func (d *Digest) Sum32() uint32 {
	l := d.length
	if l%BlockSize != 0 {
		d.Write(zeroData[l%BlockSize:])
	}
	return final(&d.h, l)
}

// Sums32 returns p's digest directly.
// len(p) must > 0. TODO return 0 when len is 0?
func Sum32(p []byte) uint32 {

	if len(p) == 0 {
		return 0
	}

	return sum32(p)
}

func (d *Digest) Reset() {
	d.remain = 0
	d.length = 0
	copy(d.h[:], initState[:])
	d.buf = [BlockSize]byte{}
}

func (d *Digest) Size() int { return Size }

func (d *Digest) BlockSize() int { return BlockSize }
