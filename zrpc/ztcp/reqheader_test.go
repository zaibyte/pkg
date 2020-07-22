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
	"os"
	"reflect"
	"testing"

	"github.com/zaibyte/pkg/xlog/xlogtest"
)

func TestMain(m *testing.M) {
	xlogtest.New()
	code := m.Run()
	xlogtest.Close()
	os.Exit(code)
}

func TestRequstHeaderCanBeEncodedAndDecoded(t *testing.T) {
	r := requestHeader{
		method: objType,
		size:   1024,
		crc:    1000,
	}
	buf := make([]byte, requestHeaderSize)
	result := r.encode(buf)
	if len(result) != requestHeaderSize {
		t.Fatalf("unexpected size")
	}
	rr := requestHeader{}
	if !rr.decode(result) {
		t.Fatalf("decode failed")
	}
	if !reflect.DeepEqual(&r, &rr) {
		t.Errorf("request header changed")
	}
}

func TestRequestHeaderCRCIsChecked(t *testing.T) {
	r := requestHeader{
		method: objType,
		size:   1024,
		crc:    1000,
	}
	buf := make([]byte, requestHeaderSize)
	result := r.encode(buf)
	if len(result) != requestHeaderSize {
		t.Fatalf("unexpected size")
	}
	rr := requestHeader{}
	if !rr.decode(result) {
		t.Fatalf("decode failed")
	}
	crc := binary.BigEndian.Uint32(result[18:])
	binary.BigEndian.PutUint32(result[18:], crc+1)
	if rr.decode(result) {
		t.Fatalf("crc error not reported")
	}
	binary.BigEndian.PutUint32(result[18:], crc)
	if !rr.decode(result) {
		t.Fatalf("decode failed")
	}
	binary.BigEndian.PutUint64(result[2:], 0)
	if rr.decode(result) {
		t.Fatalf("crc error not reported")
	}
}

func TestInvalidMethodNameIsReported(t *testing.T) {
	r := requestHeader{
		method: 1024,
		size:   1024,
		crc:    1000,
	}
	buf := make([]byte, requestHeaderSize)
	result := r.encode(buf)
	if len(result) != requestHeaderSize {
		t.Fatalf("unexpected size")
	}
	rr := requestHeader{}
	if rr.decode(result) {
		t.Fatalf("decode did not report invalid method name")
	}
}
