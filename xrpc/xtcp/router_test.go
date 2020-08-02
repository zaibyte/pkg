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

package xtcp

import (
	"math/rand"
	"testing"

	"github.com/templexxx/tsc"
)

func TestRouter_AddFunc(t *testing.T) {
	r := NewRouter()

	has := make([]uint8, 127)
	rand.Seed(tsc.UnixNano())
	allInt := rand.Perm(256)
	for i := 0; i < 127; i++ {
		has[i] = uint8(allInt[i])
	}

	for i := range has {
		if has[i] == 0 {
			continue
		}
		err := r.AddFunc(has[i], bytesHandlerFunc)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := range has {
		if has[i] == 0 {
			continue
		}
		if !r.methods.has(has[i]) {
			t.Fatal("method should have", has[i])
		}
	}
}
