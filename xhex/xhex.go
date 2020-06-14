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

// Package xhex implements hexadecimal encoding and decoding.
// xhex use AVX2 (if has) to accelerate encoding&decoding.
//
// Compare with standard lib (encoding):
// benchmark                  old ns/op     new ns/op     delta
// BenchmarkEncode/16-8       30.7          5.86          -80.91%
// BenchmarkEncode/24-8       43.4          17.8          -58.99%
// BenchmarkEncode/1024-8     1793          62.8          -96.50%
//
// benchmark                  old MB/s     new MB/s     speedup
// BenchmarkEncode/16-8       520.44       2732.67      5.25x
// BenchmarkEncode/24-8       552.44       1349.15      2.44x
// BenchmarkEncode/1024-8     571.10       16298.50     28.54x
//
// benchmark                  old ns/op     new ns/op     delta
// BenchmarkDecode/32-8       59.8          10.4          -82.61%
// BenchmarkDecode/48-8       87.5          35.3          -59.66%
// BenchmarkDecode/2048-8     3634          182           -94.99%
//
// benchmark                  old MB/s     new MB/s     speedup
// BenchmarkDecode/32-8       534.90       3074.74      5.75x
// BenchmarkDecode/48-8       548.75       1359.05      2.48x
// BenchmarkDecode/2048-8     563.56       11227.56     19.92x
// TODO decode
// Warn:
// This lib is lacking of parameters checking.
// Be sure pass legal parameters.
package xhex

import (
	"errors"
	"fmt"
)

const hextable = "0123456789abcdef"

// Encode encodes src into (2 * len(src)) bytes of dst.
//
// Warn:
// dst should have enough space(2 * len(src)),
// and len(src) must not be 0.
func Encode(dst, src []byte) {
	encode(dst, src)
}

// Define encode as a variable for reducing branch (test has AVX2 or not),
// see xhex_amd64.go for details.
var encode = func(dst, src []byte) {
	encodeBase(dst, src)
}

// encodeBase encodes src byte by byte.
func encodeBase(dst, src []byte) {
	j := 0
	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f]
		j += 2
	}
}

// ErrLength reports an attempt to decode an odd-length input
// using Decode or DecodeString.
// The stream-based Decoder returns io.ErrUnexpectedEOF instead of ErrLength.
var ErrLength = errors.New("encoding/hex: odd length hex string")

// InvalidByteError values describe errors resulting from an invalid byte in a hex string.
type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("encoding/hex: invalid byte: %#U", rune(e))
}

// Decode decodes src into len(src)/2 bytes.
//
// Decode expects that src contains only hexadecimal
// characters and that src has even length.
func Decode(dst, src []byte) error {
	return decode(dst, src)
}

var decode = func(dst, src []byte) error {
	return decodeBase(dst, src)
}

func decodeBase(dst, src []byte) error {
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		a, ok := fromHexChar(src[j-1])
		if !ok {
			return InvalidByteError(src[j-1])
		}
		b, ok := fromHexChar(src[j])
		if !ok {
			return InvalidByteError(src[j])
		}
		dst[i] = (a << 4) | b
		i++
	}
	if len(src)%2 == 1 {
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if _, ok := fromHexChar(src[j-1]); !ok {
			return InvalidByteError(src[j-1])
		}
		return ErrLength
	}
	return nil
}

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
