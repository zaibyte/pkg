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
// Copyright (c) 2016 Caleb Spare
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// This file is copied from Caleb Spare xxhash Go implementation.

package zdigest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  uint32
	}{
		{"empty", "", 0x51d8e999},
		{"a", "a", 0xa98c6e5b},
		{"as", "as", 0xd66be179},
		{"asd", "asd", 0x72a97393},
		{"asdf", "asdf", 0x99cea71e},
		{
			"len=63",
			// Exactly 63 characters, which exercises all code paths.
			"Call me Ishmael. Some years ago--never mind how long precisely-",
			0x70d6fd96,
		},
	} {
		for chunkSize := 1; chunkSize <= len(tt.input); chunkSize++ {
			name := fmt.Sprintf("%s,chunkSize=%d", tt.name, chunkSize)
			t.Run(name, func(t *testing.T) {
				testDigest(t, tt.input, chunkSize, tt.want)
			})
		}
		t.Run(tt.name, func(t *testing.T) { testSum(t, tt.input, tt.want) })
	}
}

func testDigest(t *testing.T, input string, chunkSize int, want uint32) {
	d := New()
	ds := New() // uses WriteString
	for i := 0; i < len(input); i += chunkSize {
		chunk := input[i:]
		if len(chunk) > chunkSize {
			chunk = chunk[:chunkSize]
		}
		n, err := d.Write([]byte(chunk))
		if err != nil || n != len(chunk) {
			t.Fatalf("Digest.Write: got (%d, %v); want (%d, nil)", n, err, len(chunk))
		}
		n, err = ds.WriteString(chunk)
		if err != nil || n != len(chunk) {
			t.Fatalf("Digest.WriteString: got (%d, %v); want (%d, nil)", n, err, len(chunk))
		}
	}
	if got := d.Sum32(); got != want {
		t.Fatalf("Digest.Sum32: got 0x%x; want 0x%x", got, want)
	}
	if got := ds.Sum32(); got != want {
		t.Fatalf("Digest.Sum32 (WriteString): got 0x%x; want 0x%x", got, want)
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], want)
	if got := d.Sum(nil); !bytes.Equal(got, b[:]) {
		t.Fatalf("Sum: got %v; want %v", got, b[:])
	}
}

func testSum(t *testing.T, input string, want uint32) {
	if got := Sum32([]byte(input)); got != want {
		t.Fatalf("Sum32: got 0x%x; want 0x%x", got, want)
	}
	if got := Sum32String(input); got != want {
		t.Fatalf("Sum32String: got 0x%x; want 0x%x", got, want)
	}
}

func TestReset(t *testing.T) {
	parts := []string{"The quic", "k br", "o", "wn fox jumps", " ov", "er the lazy ", "dog."}
	d := New()
	for _, part := range parts {
		d.Write([]byte(part))
	}
	h0 := d.Sum32()

	d.Reset()
	d.Write([]byte(strings.Join(parts, "")))
	h1 := d.Sum32()

	if h0 != h1 {
		t.Errorf("0x%x != 0x%x", h0, h1)
	}
}

var sink uint32

func TestAllocs(t *testing.T) {
	const shortStr = "abcdefghijklmnop"
	// Sum32([]byte(shortString)) shouldn't allocate because the
	// intermediate []byte ought not to escape.
	// (See https://github.com/cespare/xxhash/pull/2.)
	t.Run("Sum32", func(t *testing.T) {
		testAllocs(t, func() {
			sink = Sum32([]byte(shortStr))
		})
	})
	// Creating and using a Digest shouldn't allocate because its methods
	// shouldn't make it escape. (A previous version of New returned a
	// hash.Hash64 which forces an allocation.)
	t.Run("Digest", func(t *testing.T) {
		b := []byte("asdf")
		testAllocs(t, func() {
			d := New()
			d.Write(b)
			sink = d.Sum32()
		})
	})
}

func testAllocs(t *testing.T, fn func()) {
	t.Helper()
	if allocs := int(testing.AllocsPerRun(10, fn)); allocs > 0 {
		t.Fatalf("got %d allocation(s) (want zero)", allocs)
	}
}
