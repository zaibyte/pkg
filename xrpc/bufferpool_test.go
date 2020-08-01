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
	"sync"
	"testing"

	"github.com/zaibyte/pkg/xstrconv"

	"github.com/stretchr/testify/assert"
)

func TestBuffers(t *testing.T) {
	const dummyData = "dummy data"
	p := newBufferPool()

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				buf := p.Get()
				assert.Zero(t, buf.Len(), "Expected truncated buffer")
				assert.NotZero(t, buf.Cap(), "Expected non-zero capacity")

				buf.Write(xstrconv.ToBytes(dummyData))
				assert.Equal(t, buf.Len(), len(dummyData), "Expected buffer to contain dummy data")

				buf.Free()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
