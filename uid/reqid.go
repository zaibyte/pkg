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
	"encoding/hex"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_randID = rand.New(rand.NewSource(time.Now().UnixNano())).Uint32()
	hexEnc  = func(dst, src []byte) {
		hex.Encode(dst, src)
	}
)

var makeReqPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16+32)
		return &p
	},
}

// MakeReqID returns a request ID.
// Request ID is encoded in 128bit hex codes which is as same as Jaeger.
//
// Warn:
// Maybe not unique but it's acceptable.
func MakeReqID(boxID uint32) string {

	p := makeReqPool.Get().(*[]byte)
	b := *p

	binary.BigEndian.PutUint32(b[:4], boxID)
	binary.BigEndian.PutUint32(b[4:8], _randID)
	binary.BigEndian.PutUint32(b[8:12], atomic.LoadUint32(&ticker.ts))
	binary.BigEndian.PutUint32(b[12:16], atomic.AddUint32(&ticker.seqID, 1))

	hexEnc(b[16:48], b[:16])
	v := string(b[16:48])
	makeReqPool.Put(p)

	return v
}

// ParseReqID gets boxID & pid & time from a request ID.
func ParseReqID(reqID string) (boxID uint32, t time.Time, err error) {

	b := make([]byte, 16)
	n, err := hex.Decode(b[:16], []byte(reqID))
	if err != nil {
		return
	}
	b = b[:n]
	boxID = binary.BigEndian.Uint32(b[:4])
	t = ToTime(binary.BigEndian.Uint32(b[8:12]))
	return
}
