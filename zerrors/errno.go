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

package zerrors

import "errors"

// An Errno is an unsigned number describing an error condition.
// It implements the error interface. The zero Errno is by convention
// a non-error, so code to convert from Errno to error should use:
//	err = nil
//	if errno != 0 {
//		err = errno
//	}
//
type Errno uint16

func (e Errno) Error() string {
	if e == 0 {
		return ""
	}

	if int(e) < len(errnoStr) {
		s := errnoStr[e]
		if s != "" {
			return s
		}
	}
	return UnknownErrno.Error()
}

// ErrnoErr returns common boxed Errno values.
func ErrnoErr(e Errno) error {
	switch e {
	case 0:
		return nil
	default:
		return e
	}
}

// ErrErrno returns Errno value by error.
func ErrErrno(err error) Errno {
	if err == nil {
		return 0
	}

	for {
		err2 := errors.Unwrap(err)
		if err2 == nil {
			break
		}
		err = err2
	}

	u, ok := err.(Errno)
	if ok {
		return u
	}

	return UnknownErrno
}

const (
	UnknownErrno = Errno(65535)
)

// Error table
var errnoStr = [...]string{
	65535: "unknown error",
}
