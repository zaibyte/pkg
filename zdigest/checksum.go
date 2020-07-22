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

package zdigest

import "hash/crc32"

// CrcTbl uses Castagnoli which has better error detection characteristics than IEEE and faster.
var CrcTbl = crc32.MakeTable(crc32.Castagnoli)

// Application layer checksum, avoiding silent data corruption in header,
// checksum will be ignored only when TLS is enabled.
func Checksum(b []byte) uint32 {
	return crc32.Checksum(b, CrcTbl)
}
