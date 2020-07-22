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

	"github.com/zaibyte/pkg/zdigest"

	"github.com/zaibyte/pkg/xlog"
)

const requestHeaderSize = 26

type requestHeader struct {
	method uint16
	size   uint64
	reqid  uint64
	crc    uint32
}

const (
	// TCPName is the name of the tcp RPC module.
	TCPName        = "zai-tcp-transport"
	objType uint16 = 100
)

func (h *requestHeader) encode(buf []byte) []byte {
	if len(buf) < requestHeaderSize {
		panic("input buf too small")
	}
	binary.BigEndian.PutUint16(buf, h.method)
	binary.BigEndian.PutUint64(buf[2:], h.size)
	binary.BigEndian.PutUint64(buf[10:], h.reqid)
	binary.BigEndian.PutUint32(buf[18:], 0)
	binary.BigEndian.PutUint32(buf[22:], h.crc)
	v := zdigest.Checksum(buf[:requestHeaderSize])
	binary.BigEndian.PutUint32(buf[18:], v)
	return buf[:requestHeaderSize]
}

func (h *requestHeader) decode(buf []byte) bool {
	if len(buf) < requestHeaderSize {
		return false
	}

	reqid := binary.BigEndian.Uint64(buf[10:])

	incoming := binary.BigEndian.Uint32(buf[18:])
	binary.BigEndian.PutUint32(buf[18:], 0)
	expected := zdigest.Checksum(buf[:requestHeaderSize])
	if incoming != expected {
		xlog.ErrorID(reqid, "header crc check failed")
		return false
	}
	binary.BigEndian.PutUint32(buf[18:], incoming)
	method := binary.BigEndian.Uint16(buf)
	if method != objType {
		xlog.ErrorID(reqid, "invalid method type")
		return false
	}
	h.method = method
	h.size = binary.BigEndian.Uint64(buf[2:])
	h.crc = binary.BigEndian.Uint32(buf[22:])
	return true
}
