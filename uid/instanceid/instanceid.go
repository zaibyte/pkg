/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package instanceid provides method to generate unique Instance ID.
// Any ZBuffer and ZStore instance need an unique Instance ID within a box,
// and it MUST not be changed in its lifecycle.
package instanceid

import (
	"encoding/hex"

	"github.com/google/uuid"
)

// Get gets Instance ID (according MAC address).
// Only for showing a way to generate unique instance ID and testing.
//
// Warn:
// It maybe not unique when:
// 1. In container, MAC address maybe not unique in cluster.
//   (There are still ways to make it unique, try to make it and things will get easier)
// 2. After replacing a new net interface, the MAC address won't be as same as before.
//   (You should offline instance first if hardware need to be replaced.)
func Get() string {
	return hex.EncodeToString(uuid.NodeID())
}
