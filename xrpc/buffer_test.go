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

package xrpc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zaibyte/pkg/xstrconv"

	"github.com/stretchr/testify/assert"
)

func TestBufferWrites(t *testing.T) {
	buf := newBufferPool().Get()

	tests := []struct {
		desc string
		f    func()
		want string
	}{
		{"AppendWrite", func() { buf.Write([]byte("foo")) }, "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			buf.Reset()
			tt.f()
			assert.Equal(t, tt.want, string(buf.Bytes()), "Unexpected string(buffer.Bytes()).")
			assert.Equal(t, len(tt.want), buf.Len(), "Unexpected buffer length.")
			// We're not writing more than a kibibyte in tests.
			assert.Equal(t, _size, buf.Cap(), "Expected buffer capacity to remain constant.")
		})
	}
}

func BenchmarkBuffers(b *testing.B) {
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("a", 1024)
	slice := make([]byte, 1024)
	buf := bytes.NewBuffer(slice)
	custom := newBufferPool().Get()
	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice = append(slice, str...)
			slice = slice[:0]
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.WriteString(str)
			buf.Reset()
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom.Write(xstrconv.ToBytes(str))
			custom.Reset()
		}
	})
}
