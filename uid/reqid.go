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
	"sync"
	"time"

	"github.com/templexxx/tsc"
	"github.com/templexxx/xhex"

	"github.com/zaibyte/pkg/xstrconv"
)

// reqid struct:
// +-----------+------------+-----------------+---------------+
// | boxID(10) | padding(6) |  instanceID(48) | timestamp(64) |
// +-----------+------------+-----------------+---------------+
//
// Total length: 16B.
//
// boxID: 10bit
// padding: 6bit
// instanceID: 48bit
// timestamp: 64bit
//
// Because timestamp's precision is nanosecond,
// and getting timestamp has cost too,
// so it's impossible to find two same reqid
// even multi apps run on the same machine(has same instanceID).

var _instanceID = instanceIdToUint64()

const instanceIDMask = (1 << 48) - 1

func instanceIdToUint64() uint64 {
	b := makeInstanceIDBytes()
	p := make([]byte, 8)
	copy(p[:6], b)
	return binary.LittleEndian.Uint64(p)
}

// Buf will escape to heap because can't inline hex encoding.
// So make a pool here.
var reqMPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16+32)
		return &p
	},
}

// MakeReqID makes a request ID.
// Request ID is encoded in 128bit hex codes which is as same as Jaeger.
//
// Warn:
// Maybe not unique but it's acceptable.
func MakeReqID(boxID uint32) string {

	p := reqMPool.Get().(*[]byte)
	b := *p

	binary.LittleEndian.PutUint64(b[:8], uint64(boxID)<<54|_instanceID)
	binary.LittleEndian.PutUint64(b[8:16], uint64(tsc.UnixNano()))

	xhex.Encode(b[16:16+32], b[:16])
	v := string(b[16 : 16+32])
	reqMPool.Put(p)

	return v
}

// Buf will escape to heap because can't inline hex encoding.
// So make a pool here.
var reqPPool = sync.Pool{
	New: func() interface{} {
		p := make([]byte, 16)
		return &p
	},
}

// ParseReqID parses reqID.
func ParseReqID(reqID string) (boxID uint32, instanceID string, t time.Time, err error) {

	p := reqPPool.Get().(*[]byte)
	b := *p

	err = xhex.Decode(b[:16], xstrconv.ToBytes(reqID))
	if err != nil {
		reqPPool.Put(p)
		return
	}
	bi := binary.LittleEndian.Uint64(b[:8])
	boxID = uint32(bi >> 54)
	instanceID = hex.EncodeToString(b[0:6])
	t = time.Unix(0, int64(binary.LittleEndian.Uint64(b[8:16])))
	reqPPool.Put(p)

	return
}
