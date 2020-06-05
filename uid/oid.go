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
	"encoding/base64"
	"encoding/binary"
	"sync"
	"time"
)

const (
	// epoch is an Unix time.
	// 2020-06-03T08:39:34.000+0800.
	epoch = 1591144774
	// endUT is the max Unix time.
	// It will reach the end after 136 years from epoch.
	endUT = 5880040774 // epoch + 136 years.
)

// oidMPool is OID Marshal pool.
var oidMPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 24+32) // 24 for raw bytes slice, 32 for encoded.
		return &p
	},
}

// oidUPool is OID Unmarshal pool.
var oidUPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 24)
		return &p
	},
}

// OID is the unique Object ID.
// It's Globally unique (across boxes).
type OID struct {
	BoxID    uint32
	ExtentID uint32
	Digest   uint64
	Size     uint32
	TS       int64 // TS is a unix time.
}

var _oEnc = base64.URLEncoding

// Marshal returns the encoding of v.
func (o *OID) Marshal() string {

	now := time.Now().Unix()
	if now > endUT {
		panic("zai met its doom")
	}
	delta := uint32(now - epoch)

	p := oidMPool.Get().(*[]byte)

	b := *p

	binary.LittleEndian.PutUint32(b[:4], o.BoxID)
	binary.LittleEndian.PutUint32(b[4:8], o.ExtentID)
	binary.LittleEndian.PutUint64(b[8:16], o.Digest)
	binary.LittleEndian.PutUint32(b[16:20], o.Size)
	binary.LittleEndian.PutUint32(b[20:24], delta)

	_oEnc.Encode(b[24:], b[:24])
	v := string(b[24:])

	oidMPool.Put(p)

	return v
}

// Unmarshal parses the encoded data and stores the result
func (o *OID) Unmarshal(oid string) error {

	p := oidUPool.Get().(*[]byte)
	defer oidUPool.Put(p)

	b := *p

	n, err := _oEnc.Decode(b[:24], []byte(oid))
	if err != nil {
		return err
	}
	b = b[:n]

	o.BoxID = binary.LittleEndian.Uint32(b[:4])
	o.ExtentID = binary.LittleEndian.Uint32(b[4:8])
	o.Digest = binary.LittleEndian.Uint64(b[8:16])
	o.Size = binary.LittleEndian.Uint32(b[16:20])
	o.TS = int64(binary.LittleEndian.Uint32(b[20:24]) + epoch)
	return nil
}
