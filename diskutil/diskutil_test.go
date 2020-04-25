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

package diskutil

import (
	"syscall"
	"testing"

	"github.com/zaibyte/pkg/xerrors"
)

func TestIsBroken(t *testing.T) {
	err := syscall.EIO
	if !IsBroken(err) {
		t.Fatal("should be broken")
	}
	xerr := xerrors.WithMsg(err, "EIO")
	if !IsBroken(xerr) {
		t.Fatal("should be broken")
	}

	err = syscall.EROFS
	if !IsBroken(err) {
		t.Fatal("should be broken")
	}
	xerr = xerrors.WithMsg(err, "EROFS")
	if !IsBroken(xerr) {
		t.Fatal("should be broken")
	}
}
