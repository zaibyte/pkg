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
// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file contains code derived from Dragonboat.
// The main logic & codes are copied from Dragonboat.

package ztcp

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zaibyte/pkg/uid"
)

func TestRequestHeaderCanBeEncodedAndDecoded(t *testing.T) {
	r := requestHeader{
		msgID:  2048,
		method: objGetMethod,
		reqid:  uid.MakeReqID(),
		size:   1024,
		crc:    1000,
	}
	buf := make([]byte, requestHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, requestHeaderSize, len(result))

	rr := requestHeader{}
	assert.Nil(t, rr.decode(result))

	assert.Equal(t, r, rr)
}

func TestRequestHeaderCRCIsChecked(t *testing.T) {
	r := requestHeader{
		msgID:  2048,
		method: objGetMethod,
		reqid:  uid.MakeReqID(),
		size:   1024,
		crc:    1000,
	}
	buf := make([]byte, requestHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, requestHeaderSize, len(result))

	rr := requestHeader{}
	assert.Nil(t, rr.decode(result))

	crc := binary.BigEndian.Uint32(result[26:])
	binary.BigEndian.PutUint32(result[26:], crc+1)
	assert.Equal(t, ErrChecksumMismatch, rr.decode(result))

	binary.BigEndian.PutUint32(result[26:], crc)
	assert.Nil(t, rr.decode(result))

	binary.BigEndian.PutUint64(result[2:], 0)
	assert.Equal(t, ErrChecksumMismatch, rr.decode(result))
}

func TestInvalidMethodNameIsReported(t *testing.T) {

	methods := []uint16{0, maxMethod + 1}

	for _, method := range methods {
		r := requestHeader{
			msgID:  2048,
			reqid:  uid.MakeReqID(),
			method: method,
			size:   1024,
			crc:    1000,
		}
		buf := make([]byte, requestHeaderSize)
		result := r.encode(buf)
		assert.Equal(t, requestHeaderSize, len(result))

		rr := requestHeader{}
		assert.Equal(t, ErrInvalidMethod, rr.decode(result))
	}
}

func TestRespHeaderCanBeEncodedAndDecoded(t *testing.T) {
	r := respHeader{
		msgID: 2048,
		size:  1024,
		crc:   1000,
	}
	buf := make([]byte, respHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, respHeaderSize, len(result))

	rr := respHeader{}
	assert.Nil(t, rr.decode(result))

	assert.Equal(t, r, rr)
}

func TestRespHeaderCRCIsChecked(t *testing.T) {
	r := respHeader{
		msgID: 2048,
		size:  1024,
		crc:   1000,
	}
	buf := make([]byte, respHeaderSize)
	result := r.encode(buf)
	assert.Equal(t, respHeaderSize, len(result))

	rr := respHeader{}
	assert.Nil(t, rr.decode(result))

	crc := binary.BigEndian.Uint32(result[16:])
	binary.BigEndian.PutUint32(result[16:], crc+1)
	assert.Equal(t, ErrChecksumMismatch, rr.decode(result))

	binary.BigEndian.PutUint32(result[16:], crc)
	assert.Nil(t, rr.decode(result))

	binary.BigEndian.PutUint64(result[2:], 0)
	assert.Equal(t, ErrChecksumMismatch, rr.decode(result))
}
