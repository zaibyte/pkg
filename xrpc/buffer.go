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

import "github.com/zaibyte/pkg/config/settings"

const _size = 32 * 1024                     // by default, create 32 KiB buffers
const _bigsize = 1024 + settings.MaxObjSize // 1024 for main request. MaxObjSize for avoiding slice expansion.

// Buffer is a thin wrapper around a byte slice. It's intended to be pooled, so
// the only way to construct one is via a bufferPool.
type Buffer struct {
	BS   []byte
	pool bufferPool
}

// Len returns the length of the underlying byte slice.
func (b *Buffer) Len() int {
	return len(b.BS)
}

// Cap returns the capacity of the underlying byte slice.
func (b *Buffer) Cap() int {
	return cap(b.BS)
}

// Bytes returns a mutable reference to the underlying byte slice.
func (b *Buffer) Bytes() []byte {
	return b.BS
}

// Reset resets the underlying byte slice. Subsequent writes re-use the slice's
// backing array.
func (b *Buffer) Reset() {
	b.BS = b.BS[:0]
}

// Write implements io.Writer.
func (b *Buffer) Write(bs []byte) (int, error) {
	b.BS = append(b.BS, bs...)
	return len(bs), nil
}

// Free returns the Buffer to its bufferPool.
//
// Callers must not retain references to the Buffer after calling Free.
func (b *Buffer) Free() {
	b.pool.put(b)
}

// MarshalTo returns byte slice inside Buffer directly,
// for avoiding memory copy.
//
// The data must haven been written into Buffer.
func (b *Buffer) MarshalTo(_ []byte) ([]byte, error) {
	return b.Bytes(), nil
}

func (b *Buffer) MarshalLen() (int, error) {
	return b.Len(), nil
}

// UnmarshalBinary do nothing just for satisfying Marshaler interface.
//
// The data must haven been written into Buffer.
func (b *Buffer) UnmarshalBinary(_ []byte) error {

	return nil
}
