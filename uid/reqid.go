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

// Package reqid provides functions to generate unique Request ID.
package uid

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"sync"
	"time"
)

// default max_pid = num_processors * 1024,
// or max_pid = 32768 when num_processors < 32.
// Uint16 may not enough, so uint32.
var _pid = uint32(os.Getpid())

var _reqEnc = base64.URLEncoding

var makeReqPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16+24) // 16 for raw bytes slice, 24 for encoded.
		return &p
	},
}

var parseReqPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16)
		return &p
	},
}

// MakeReqID returns a request ID.
// warn: maybe not unique but it's acceptable.
func MakeReqID(boxID uint32) string {
	return MakeReqIDWithTime(boxID, time.Now())
}

// MakeReqIDWithTime returns a request ID with specific time.
// warn: maybe not unique but it's acceptable.
func MakeReqIDWithTime(boxID uint32, t time.Time) string {

	p := makeReqPool.Get().(*[]byte)

	b := *p

	binary.LittleEndian.PutUint32(b[:], boxID)
	binary.LittleEndian.PutUint32(b[4:8], _pid)
	binary.LittleEndian.PutUint64(b[8:16], uint64(t.UnixNano()))

	_reqEnc.Encode(b[16:], b[:16])
	v := string(b[16:])

	makeReqPool.Put(p)

	return v
}

// ParseReqID gets boxID & pid & time from a request ID.
func ParseReqID(reqID string) (boxID uint32, pid uint32, t time.Time, err error) {

	p := parseReqPool.Get().(*[]byte)
	defer parseReqPool.Put(p)

	b := *p

	n, err := _reqEnc.Decode(b[:16], []byte(reqID))
	if err != nil {
		return
	}
	b = b[:n]
	boxID = binary.LittleEndian.Uint32(b[:4])
	pid = binary.LittleEndian.Uint32(b[4:8])
	t = time.Unix(0, int64(binary.LittleEndian.Uint64(b[8:16])))
	return
}
