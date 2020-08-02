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

import "sync"

var (
	_bufferPool = newBufferPool()
	// GetBytes retrieves a buffer from the buffer pool, creating one if necessary.
	GetBytes = _bufferPool.Get

	_bigbufferPool = newBigBufferPool()
	GetBigBytes    = _bigbufferPool.Get()
)

// A bufferPool is a type-safe wrapper around a sync.bufferPool.
type bufferPool struct {
	p *sync.Pool
}

// newBufferPool constructs a new bufferPool.
func newBufferPool() bufferPool {
	return bufferPool{p: &sync.Pool{
		New: func() interface{} {
			return &Buffer{BS: make([]byte, 0, _size)}
		},
	}}
}

// newBigBufferPool constructs a new bufferPool with a bigger buffer.
func newBigBufferPool() bufferPool {
	return bufferPool{p: &sync.Pool{
		New: func() interface{} {
			return &Buffer{BS: make([]byte, 0, _bigsize)}
		},
	}}
}

// Get retrieves a Buffer from the pool, creating one if necessary.
func (p bufferPool) Get() *Buffer {
	buf := p.p.Get().(*Buffer)
	buf.Reset()
	buf.pool = p
	return buf
}

func (p bufferPool) put(buf *Buffer) {
	p.p.Put(buf)
}
