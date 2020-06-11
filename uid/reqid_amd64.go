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
// hecEncAVX2's algorithm is copied from https://github.com/zbjornson/fast-hex.
// Copyright (c) 2017 Zach Bjornson

package uid

import (
	"github.com/templexxx/cpu"
)

func init() {
	if cpu.X86.HasAVX2 {
		hexEnc = func(dst, src []byte) {
			hexEncAVX2(&dst[0], &src[0])
		}
	}
}

// hextable doubles normal hex table for AVX2 register (expand 16Bytes to 32Bytes)
var hextable = "0123456789abcdef0123456789abcdef"

// lows is used for generating bytes' low part.
// Magic from Zach Bjornson's contribution.
var lows = []int8{-1, 0, -1, 2, -1, 4, -1, 6, -1, 8, -1, 10, -1, 12, -1, 14,
	-1, 0, -1, 2, -1, 4, -1, 6, -1, 8, -1, 10, -1, 12, -1, 14}

// hexEncAVX2 implements hexadecimal encoding,
// encodes src into dst.
// Warn:
// It's not general purpose hexadecimal encoding,
// for reqid we only which mean len(src) must be 16.
// TODO: remove the len limit.
//go:noescape
func hexEncAVX2(dst, src *byte)
