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
package reqid

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"time"
)

// default max_pid = num_processors * 1024,
// or max_pid = 32768 when num_processors < 32.
// Uint16 may not enough, so uint32.
var _pid = uint32(os.Getpid())

// Next returns a request ID.
// warn: maybe not unique but it's acceptable.
func Next() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], _pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

// Parse gets pid & time from a reqID.
func Parse(reqID string) (pid uint32, t time.Time, err error) {

	b, err := base64.URLEncoding.DecodeString(reqID)
	if err != nil {
		return
	}

	pid = binary.LittleEndian.Uint32(b[:4])
	nt := int64(binary.LittleEndian.Uint64(b[4:]))
	t = time.Unix(0, nt)
	return
}
