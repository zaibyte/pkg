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

package xtcp

import (
	"encoding/binary"

	"github.com/zaibyte/pkg/xrpc"

	"github.com/zaibyte/pkg/xdigest"
)

const reqHeaderSize = 25 + 16

// reqHeader is the header for request.
type reqHeader struct {
	method   uint8    // [0, 1)
	msgID    uint64   // [1, 9)
	reqid    uint64   // [9, 17)
	bodySize uint32   // [17, 21)
	oid      [16]byte // [21, 37)
	crc      uint32   // [37, 41)
}

func (h *reqHeader) encode(buf []byte) []byte {
	if len(buf) < reqHeaderSize {
		panic("input buf too small")
	}
	buf[0] = h.method
	binary.BigEndian.PutUint64(buf[1:9], h.msgID)
	binary.BigEndian.PutUint64(buf[9:17], h.reqid)
	binary.BigEndian.PutUint32(buf[17:21], h.bodySize)
	copy(buf[21:37], h.oid[:])
	binary.BigEndian.PutUint32(buf[37:41], 0)
	crc := xdigest.Checksum(buf[:reqHeaderSize])
	binary.BigEndian.PutUint32(buf[37:41], crc)
	h.crc = crc
	return buf[:reqHeaderSize]
}

func (h *reqHeader) decode(buf []byte) error {
	if len(buf) < reqHeaderSize {
		panic("input buf too small")
	}

	incoming := binary.BigEndian.Uint32(buf[37:41])
	binary.BigEndian.PutUint32(buf[37:41], 0)
	expected := xdigest.Checksum(buf[:reqHeaderSize])
	if incoming != expected {
		return xrpc.ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[37:41], incoming)

	h.method = buf[0]
	h.msgID = binary.BigEndian.Uint64(buf[1:9])
	h.reqid = binary.BigEndian.Uint64(buf[9:17])
	h.bodySize = binary.BigEndian.Uint32(buf[17:21])
	copy(h.oid[:], buf[21:37])
	h.crc = incoming

	return nil
}

const respHeaderSize = 18

// respHeader is the header for response.
type respHeader struct {
	msgID uint64 // [0, 8)
	errno uint16 // [8, 10)
	size  uint32 // [10, 14)
	crc   uint32 // [14, 18)
}

func (h *respHeader) encode(buf []byte) []byte {
	if len(buf) < respHeaderSize {
		panic("input buf too small")
	}
	binary.BigEndian.PutUint64(buf[0:8], h.msgID)
	binary.BigEndian.PutUint16(buf[8:10], h.errno)
	binary.BigEndian.PutUint32(buf[10:14], h.size)
	binary.BigEndian.PutUint32(buf[14:18], 0)
	crc := xdigest.Checksum(buf[:respHeaderSize])
	binary.BigEndian.PutUint32(buf[14:18], crc)
	h.crc = crc
	return buf[:respHeaderSize]
}

func (h *respHeader) decode(buf []byte) error {
	if len(buf) < respHeaderSize {
		panic("input buf too small")
	}

	incoming := binary.BigEndian.Uint32(buf[14:18])
	binary.BigEndian.PutUint32(buf[14:18], 0)
	expected := xdigest.Checksum(buf[:respHeaderSize])
	if incoming != expected {
		return xrpc.ErrChecksumMismatch
	}
	binary.BigEndian.PutUint32(buf[14:18], incoming)

	h.msgID = binary.BigEndian.Uint64(buf[0:8])
	h.errno = binary.BigEndian.Uint16(buf[8:10])
	h.size = binary.BigEndian.Uint32(buf[10:14])
	h.crc = incoming
	return nil
}
