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
)

const requestHeaderSize = 34

type requestHeader struct {
	method uint16 // [0, 2)
	msgID  uint64 // [2, 10)
	size   uint64 // [10, 18)
	reqid  uint64 // [18, 26)
	hcrc   uint32 // [26, 30), Header CRC.
	crc    uint32 // [30, 34)
}

// method: [1, 256)
const (
	maxMethod uint16 = 255 // method must <= maxMethod.

	objPutMethod uint16 = 1
	objGetMethod uint16 = 2
	objDelMethod uint16 = 3
)

func (h *requestHeader) encode(buf []byte) []byte {
	if len(buf) < requestHeaderSize {
		panic("input buf too small")
	}
	binary.BigEndian.PutUint16(buf[0:2], h.method)
	binary.BigEndian.PutUint64(buf[2:10], h.msgID)
	binary.BigEndian.PutUint64(buf[10:18], h.size)
	binary.BigEndian.PutUint64(buf[18:26], h.reqid)
	binary.BigEndian.PutUint32(buf[26:30], 0)
	binary.BigEndian.PutUint32(buf[30:34], h.crc)
	v := zdigest.Checksum(buf[:requestHeaderSize])
	binary.BigEndian.PutUint32(buf[26:30], v)
	return buf[:requestHeaderSize]
}

func (h *requestHeader) decode(buf []byte) error {
	if len(buf) < requestHeaderSize {
		return ErrBadMessage
	}

	reqid := binary.BigEndian.Uint64(buf[18:26])
	h.reqid = reqid

	incoming := binary.BigEndian.Uint32(buf[26:30])
	binary.BigEndian.PutUint32(buf[26:30], 0)
	expected := zdigest.Checksum(buf[:requestHeaderSize])
	if incoming != expected {
		return ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[26:30], incoming)
	method := binary.BigEndian.Uint16(buf)
	if method == 0 || method > maxMethod {
		return ErrInvalidMethod
	}
	h.method = method
	h.msgID = binary.BigEndian.Uint64(buf[2:10])
	h.size = binary.BigEndian.Uint64(buf[10:18])
	h.crc = binary.BigEndian.Uint32(buf[30:34])
	return nil
}

const respHeaderSize = 24

type respHeader struct {
	msgID uint64 // [0, 8)
	size  uint64 // [8, 16)
	hcrc  uint32 // [16, 20), Header CRC.
	crc   uint32 // [20,24)
}

func (h *respHeader) encode(buf []byte) []byte {
	if len(buf) < respHeaderSize {
		panic("input buf too small")
	}
	binary.BigEndian.PutUint64(buf[0:8], h.msgID)
	binary.BigEndian.PutUint64(buf[8:16], h.size)
	binary.BigEndian.PutUint32(buf[16:20], 0)
	binary.BigEndian.PutUint32(buf[20:24], h.crc)
	v := zdigest.Checksum(buf[:respHeaderSize])
	binary.BigEndian.PutUint32(buf[16:20], v)
	return buf[:respHeaderSize]
}

func (h *respHeader) decode(buf []byte) error {
	if len(buf) < respHeaderSize {
		return ErrBadMessage
	}

	incoming := binary.BigEndian.Uint32(buf[16:20])
	binary.BigEndian.PutUint32(buf[16:20], 0)
	expected := zdigest.Checksum(buf[:respHeaderSize])
	if incoming != expected {
		return ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[16:20], incoming)

	h.msgID = binary.BigEndian.Uint64(buf[0:8])
	h.size = binary.BigEndian.Uint64(buf[8:16])
	h.crc = binary.BigEndian.Uint32(buf[20:24])
	return nil
}
