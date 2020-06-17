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
	"sync/atomic"
	"time"

	"github.com/templexxx/xhex"

	"github.com/zaibyte/pkg/xstrconv"
)

const (
	NormalObj uint8 = 1 // Normal Object, maximum size is 4MB.
	LinkObj   uint8 = 2 // Link Object, maximum size 4MB.
)

var oidMPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 24+48, 24+48) // 24 for raw bytes slice, 48 for hex.
		return &p
	},
}

// MakeOID makes oid.
func MakeOID(boxID, extID uint32, digest uint64, size uint32, otype uint8) string {

	p := oidMPool.Get().(*[]byte)
	b := *p

	binary.BigEndian.PutUint32(b[:4], boxID)
	binary.BigEndian.PutUint32(b[4:8], extID)

	binary.BigEndian.PutUint32(b[8:12], atomic.LoadUint32(&ticker.ts))
	binary.BigEndian.PutUint64(b[12:20], digest)

	so := uint32(otype)<<24 | size // The maximum size of an object is 4MB (2^22 < 2 ^24).
	binary.BigEndian.PutUint32(b[20:24], so)

	xhex.Encode(b[24:24+48], b[:24])
	v := string(b[24 : 24+48])
	oidMPool.Put(p)

	return v
}

var oidPPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 24, 24)
		return &p
	},
}

// ParseOID parses oid.
func ParseOID(oid string) (boxID, extID uint32, t time.Time, digest uint64, size uint32, otype uint8, err error) {

	p := oidPPool.Get().(*[]byte)
	b := *p

	err = xhex.Decode(b[:24], xstrconv.ToBytes(oid))
	if err != nil {
		oidPPool.Put(p)
		return
	}
	boxID = binary.BigEndian.Uint32(b[:4])
	extID = binary.BigEndian.Uint32(b[4:8])
	t = Ts2Time(binary.BigEndian.Uint32(b[8:12]))
	digest = binary.BigEndian.Uint64(b[12:20])

	so := binary.BigEndian.Uint32(b[20:24])
	size = so << 8 >> 8
	otype = uint8(so >> 24)
	oidPPool.Put(p)
	return
}
