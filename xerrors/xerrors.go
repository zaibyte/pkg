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

// Package xerrors provides rich errors.
// Copy from github.com/pkg/errors.
package xerrors

import (
	"fmt"
	"io"
)

type withMsg struct {
	cause error
	msg   string
}

// WithMsg annotates err with a new message.
// We can get K&D style error. (http://www.gopl.io/).
//
// If err is nil, WithMsg returns nil.
func WithMsg(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &withMsg{
		cause: err,
		msg:   msg,
	}
}

func (w *withMsg) Error() string { return w.msg + ": " + w.cause.Error() }

// Unwrap return the cause of error, and implement interface {
//		Unwrap() error
//	}
// Provides compatibility for Go 1.13 error chains.
func (w *withMsg) Unwrap() error { return w.cause }

func (w *withMsg) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Unwrap())
			io.WriteString(s, w.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
