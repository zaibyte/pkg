/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uid

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/templexxx/tsc"

	"github.com/templexxx/xhex"

	"github.com/zaibyte/pkg/xstrconv"
)

// oid struct:
// +-----------+-------------+---------+------------+----------+----------+
// | boxID(10) | groupID(22) |  ts(32) | digest(32) | size(24) | otype(8) |
// +-----------+-------------+---------+------------+----------+----------+
//
// Total length: 16B. After hex encoding, it will be 32B.
//
// boxID: 10bit
// groupID: 22bit
// digest: 32bit
// ts: 32bit
// size: 24bit
// otype: 8bit

const (
	// epoch is an Unix time.
	// 2020-06-03T08:39:34.000+0800.
	epoch     int64 = 1591144774
	epochNano       = epoch * int64(time.Second)
	// doom is the zai's max Unix time.
	// It will reach the end after 136 years from epoch.
	doom int64 = 5880040774 // epoch + 136 years (about 2^32 seconds).
	// maxTS is the zai's max timestamp.
	maxTS = uint32(doom - epoch)
)

// Object types.
const (
	NormalObj uint8 = 1 // Normal Object, maximum size is 8MB.
	LinkObj   uint8 = 2 // Link Object, it links 131072 objects together (at most 1TB).
)

const groupIDMask = (1 << 22) - 1

var oidMPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16+32, 16+32) // 16 for raw bytes slice, 32 for hex.
		return &p
	},
}

// MakeOID makes oid.
// Returns ts & hex codes.
func MakeOID(boxID, groupID, digest, size uint32, otype uint8) (uint32, string) {

	ts := GetOidTS()

	return ts, MakeOIDWithTS(boxID, groupID, ts, digest, size, otype)
}

func GetOidTS() uint32 {
	now := tsc.UnixNano()
	sec := now / int64(time.Second)
	ts := uint32(sec - epoch)
	if ts >= maxTS {
		panic("zai met its doom")
	}
	return ts
}

// MakeOIDWithTS makes oid with provided ts.
// Returns hex codes.
func MakeOIDWithTS(boxID, groupID, ts, digest, size uint32, otype uint8) string {

	p := oidMPool.Get().(*[]byte)
	b := *p

	binary.LittleEndian.PutUint32(b[:4], boxID<<22|groupID)
	binary.LittleEndian.PutUint32(b[4:8], ts)

	binary.LittleEndian.PutUint32(b[8:12], digest)
	binary.LittleEndian.PutUint32(b[12:16], size<<8|uint32(otype))

	xhex.Encode(b[16:16+32], b[:16])
	v := string(b[16 : 16+32])
	oidMPool.Put(p)

	return v
}

var oidPPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16, 16)
		return &p
	},
}

// ParseOID parses oid.
func ParseOID(oid string) (boxID, groupID, ts, digest, size uint32, otype uint8, err error) {

	p := oidPPool.Get().(*[]byte)
	b := *p

	err = xhex.Decode(b[:16], xstrconv.ToBytes(oid))
	if err != nil {
		oidPPool.Put(p)
		return
	}

	boxID, groupID, ts, digest, size, otype = ParseOIDBytes(b)

	oidPPool.Put(p)

	return
}

// ParseOIDBytes parses oid in bytes.
// Assume there are enough space in oid bytes.
func ParseOIDBytes(oid []byte) (boxID, groupID, ts, digest, size uint32, otype uint8) {
	b := oid
	bg := binary.LittleEndian.Uint32(b[:4])
	boxID = bg >> 22
	groupID = bg & groupIDMask

	ts = binary.LittleEndian.Uint32(b[4:8])

	digest = binary.LittleEndian.Uint32(b[8:12])

	so := binary.LittleEndian.Uint32(b[12:16])
	size = so >> 8
	otype = uint8(so)
	return
}
