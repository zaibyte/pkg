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
	"math/rand"
	"time"

	"github.com/templexxx/tsc"
)

// reqid struct:
// +-----------+---------------+
// | randID(2) | timestamp(62) |
// +-----------+---------------+
//
// Total length: 8B (After hex encoding, it's 16B).
//
// randID: 2bit
// timestamp: 62bit
//
// Because timestamp's precision is nanosecond (details see tsc.UnixNano()),
// and getting timestamp has cost too,
// so it's almost impossible to find two same reqid with 2bit randID.

var _randID = uint64(rand.New(rand.NewSource(tsc.UnixNano())).Int63n(4))

// MakeReqID makes a request ID.
// Request ID is encoded in 64bit unsigned integer.
//
// Warn:
// Maybe not unique but it's acceptable.
func MakeReqID() uint64 {

	return _randID<<62 | (uint64(tsc.UnixNano()) - uint64(epochNano))
}

const reqTSMask = (1 << 62) - 1

// ParseReqID parses reqID.
func ParseReqID(reqID uint64) (t time.Time) {

	ts := reqID & reqTSMask
	return TsNanoToTime(ts)
}

// TsNanoToTime converts zai nanosecond ts to time.
func TsNanoToTime(ts uint64) time.Time {
	nano := int64(ts) + epochNano
	return time.Unix(0, nano)
}
