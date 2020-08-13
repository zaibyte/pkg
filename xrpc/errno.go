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

package xrpc

import "errors"

// An Errno is an unsigned number describing an error condition.
// It implements the error interface. The zero Errno is by convention
// a non-error, so code to convert from Errno to error should use:
//	err = nil
//	if errno != 0 {
//		err = errno
//	}
//
// Errno provides error numbers for indicating errors.
// Saving network I/O and marshal & unmarshal cost in RPC between Zai and ZBuf/ZCold.
//
// We don't need to support all error types, because any error should be
// logged in where it raises. For the client, it just need
// to know there is an error in the request, and what the type it is.
//
// There are two major types of errno:
// 1. Server side
// Not found error & internal server error.
// Not found: means there is no need to retry, so it's important.
// Other errors could be combined as internal server error.
//
// 2. Client side
// Bad request, not implemented*, canceled, timeout, too many request*, connection error
// Bad request: could happen when there is an illegal request.
// Not implemented: request a method which is not found.
// For saving network cost, the method will be checked in client side.
// (Client created by xtcp.Router will check method)
// Connection error means network issues.
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
	return "unknown error"
}

func (e Errno) ToErr() error {
	if e == 0 {
		return nil
	}

	return e
}

// ErrToErrno returns Errno value by error.
func ErrToErrno(err error) Errno {
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

	return Errno(internalServerError)
}

const (
	badRequest           = 1
	notFound             = 2
	notImplemented       = 3
	timeout              = 4
	tooManyRequests      = 5
	internalServerError  = 6
	connectionError      = 7
	canceled             = 8
	checksumMismatch     = 9
	invalidMethod        = 10
	requestQueueOverflow = 11
	timeBackwards        = 12
	extentFull           = 13
	diskWriteStall       = 15
)

// Error table.
// Please add errno in order.
var errnoStr = [...]string{
	badRequest:           "bad message",
	notFound:             "not found",
	notImplemented:       "not implemented",
	timeout:              "timeout",
	tooManyRequests:      "too many requests",
	internalServerError:  "internal server error",
	connectionError:      "connection error",
	canceled:             "canceled",
	checksumMismatch:     "checksum mismatch",
	invalidMethod:        "invalid method",
	requestQueueOverflow: "request queue overflow",
	timeBackwards:        "time gone backwards",
	extentFull:           "extent is full",
	diskWriteStall:       "disk write stall",
}

var (
	ErrBadRequest           = Errno(badRequest)
	ErrNotFound             = Errno(notFound) // When server side raises a not found error, using this variable.
	ErrNotImplemented       = Errno(notImplemented)
	ErrTimeout              = Errno(timeout)
	ErrTooManyRequests      = Errno(tooManyRequests)
	ErrInternalServer       = Errno(internalServerError)
	ErrConnection           = Errno(connectionError)
	ErrCanceled             = Errno(canceled)
	ErrChecksumMismatch     = Errno(checksumMismatch)
	ErrInvalidMethod        = Errno(invalidMethod)
	ErrRequestQueueOverflow = Errno(requestQueueOverflow)
	ErrTimeBackwards        = Errno(timeBackwards)
	ErrExtentFull           = Errno(extentFull)
	ErrDiskWriteStall       = Errno(diskWriteStall)
)
