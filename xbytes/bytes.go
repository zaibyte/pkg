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

package xbytes

import (
	"io"
	"sync"
)

type Buffer interface {
	io.ReadWriteCloser
	Bytes() []byte
	Set(b []byte)
}

const MaxBytesSizeInPool = 32 * 1024 // by default, create 32 KiB buffers.

// A BytesBuffer implements the io.ReadCloser, io.Writer interface by reading from
// a byte slice.
type BytesBuffer struct {
	S []byte
	i int64 // current reading index
}

// Read implements the io.Reader interface.
func (r *BytesBuffer) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.S)) {
		return 0, io.EOF
	}
	n = copy(b, r.S[r.i:])
	r.i += int64(n)
	return
}

// Write implements the io.Writer interface.
func (r *BytesBuffer) Write(b []byte) (n int, err error) {
	r.S = append(r.S, b...)
	return len(b), nil
}

// Close implements the io.Closer interface.
func (r *BytesBuffer) Close() error {
	r.S = nil // Release the byte slice.
	return nil
}

// Bytes returns a mutable reference to the underlying byte slice.
// Implements Buffer.
func (r *BytesBuffer) Bytes() []byte {
	return r.S
}

// Set sets b as underlying byte slice and reset read index.
func (r *BytesBuffer) Set(b []byte) {
	r.S = b
	r.i = 0
}

var (
	_bufferPool = newBufferPool()
	// GetBytes retrieves a buffer from the buffer pool, creating one if necessary.
	GetBytes  = _bufferPool.Get
	GetNBytes = func(n int) Buffer {
		if n <= MaxBytesSizeInPool {
			return GetBytes()
		}
		return &BytesBuffer{
			S: make([]byte, 0, n),
		}
	}
)

// A bufferPool is a type-safe wrapper around a sync.bufferPool.
type bufferPool struct {
	p *sync.Pool
}

// newBufferPool constructs a new bufferPool.
func newBufferPool() bufferPool {
	return bufferPool{p: &sync.Pool{
		New: func() interface{} {
			return &BytesBufferPool{S: make([]byte, 0, MaxBytesSizeInPool)}
		},
	}}
}

// Get retrieves a BytesBufferPool from the pool, creating one if necessary.
func (p bufferPool) Get() *BytesBufferPool {
	buf := p.p.Get().(*BytesBufferPool)
	buf.reset()
	buf.pool = p
	return buf
}

func (p bufferPool) put(buf *BytesBufferPool) {
	p.p.Put(buf)
}

// BytesBufferPool is a thin wrapper around a byte slice. It's intended to be pooled, so
// the only way to construct one is via a bufferPool.
type BytesBufferPool struct {
	S    []byte
	i    int64
	pool bufferPool
}

// Write implements the io.Writer interface.
func (r *BytesBufferPool) Write(bs []byte) (int, error) {
	r.S = append(r.S, bs...)
	return len(bs), nil
}

// Read implements the io.Reader interface.
func (r *BytesBufferPool) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.S)) {
		return 0, io.EOF
	}
	n = copy(b, r.S[r.i:])
	r.i += int64(n)
	return
}

// Close returns the BytesBufferPool to its bufferPool.
//
// Callers must not retain references to the BytesBufferPool after calling Close.
func (r *BytesBufferPool) Close() error {
	r.pool.put(r)
	return nil
}

// Bytes returns a mutable reference to the underlying byte slice.
// Implements Buffer.
func (r *BytesBufferPool) Bytes() []byte {
	return r.S
}

// Set sets b as underlying byte slice and reset read index.
func (r *BytesBufferPool) Set(b []byte) {
	r.S = b
	r.i = 0
}

// reset resets the underlying byte slice. Subsequent writes re-use the slice's
// backing array.
func (r *BytesBufferPool) reset() {
	r.S = r.S[:0]
	r.i = 0
}
