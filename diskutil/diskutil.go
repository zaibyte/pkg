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

// Package diskutil implements methods to access disk status.
package diskutil

import (
	"errors"
	"syscall"
)

// IsBroken returns an error is disk error or not.
func IsBroken(err error) bool {
	if err == nil {
		return false
	}

	// EIO: I/O error
	if errors.Is(err, syscall.EIO) {
		return true
	}

	// EROFS: Read-only file system, caused by
	// 1. VFS error,
	// 2. hard disk error
	if errors.Is(err, syscall.EROFS) {
		return true
	}

	return false
}

// GetFreeSize returns disk free space size (unit: Byte).
func GetFreeSize(path string) (free uint64, err error) {
	return getFreeSpace(path)
}
