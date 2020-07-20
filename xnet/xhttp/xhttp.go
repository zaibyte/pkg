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

// package xhttp provides http implementation for Zai.
// Zai uses HTTP/1.1 for handling management requests,
// these requests are not performance sensitive,
// HTTP/1.1 is a good choice to make things easier.
//
// Compare with HTTP/1.1 in standard lib:
// 1. Add some default server/client configs.
// 2. Add E2E checksum.
// 3. Add private headers.
// 4. Wrap some basic methods/functions, make it easier to use.
package xhttp

import "strconv"

// These header names will be added in HTTP header in Zai.
const (
	ReqIDHeader    = "x-zai-request-id"
	ChecksumHeader = "x-zai-checksum"
)

func reqIDStrToInt(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}
