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
	"errors"
	"math"
	"testing"

	"github.com/zaibyte/pkg/xerrors"

	"github.com/stretchr/testify/assert"
)

func TestErrno_Error(t *testing.T) {
	assert.Equal(t, "", Errno(0).Error())
	for i := 1; i <= math.MaxUint16; i++ {
		err := Errno(i)
		if errnoStr[i] == "" {
			assert.Equal(t, "unknown error", err.Error())
		} else {
			assert.Equal(t, errnoStr[i], err.Error())
		}
	}
}

func TestErrToErrno(t *testing.T) {

	for i := 0; i < len(errnoStr); i++ {
		exp := Errno(i)
		var err error
		err = exp
		for j := 0; j < 3; j++ {
			err = xerrors.WithMessage(exp, "with msg")
		}
		assert.True(t, errors.Is(err, exp))
	}

	err := errors.New("new error")
	err = ErrToErrno(err)
	assert.True(t, errors.Is(err, Errno(internalServerError)))
}

func TestUnitErrno(t *testing.T) {
	var i uint16
	for i = 0; i < 3; i++ {
		err := Errno(i)
		ii := uint16(err)
		assert.Equal(t, i, ii)

		err2 := Errno(ii)
		assert.Equal(t, err, err2)
	}
}

func BenchmarkErrno_Error(b *testing.B) {

	for i := 0; i < b.N; i++ {
		err := Errno(uint16(i))
		_ = err.Error()
	}
}

func BenchmarkErrToErrno(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_ = ErrToErrno(Errno(i))
	}
}
