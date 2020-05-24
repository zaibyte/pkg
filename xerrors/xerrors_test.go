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
/*
 * Copyright (c) 2015, Dave Cheney <dave@cheney.net>
 * All rights reserved.

 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:

 * * Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.

 * * Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.

 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package xerrors

import (
	"errors"
	"io"
	"testing"
)

func TestWithMessageNil(t *testing.T) {
	got := WithMessage(nil, "no error")
	if got != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWithMessage(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestWithMessagefNil(t *testing.T) {
	got := WithMessagef(nil, "no error")
	if got != nil {
		t.Errorf("WithMessage(nil, \"no error\"): got %#v, expected nil", got)
	}
}

func TestWithMessagef(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessagef(io.EOF, "read error without format specifier"), "client error", "client error: read error without format specifier: EOF"},
		{WithMessagef(io.EOF, "read error with %d format specifier", 1), "client error", "client error: read error with 1 format specifier: EOF"},
	}

	for _, tt := range tests {
		got := WithMessagef(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
		}
	}
}

func TestWithMessageIs(t *testing.T) {

	target := io.EOF

	tests := []struct {
		err     error
		message string
	}{
		{target, "read error"},
		{WithMessagef(target, "read error without format specifier"), "client error"},
		{WithMessagef(target, "read error with %d format specifier", 1), "client error"},
	}

	for _, tt := range tests {
		errWithMsg := WithMessagef(tt.err, tt.message)
		if !errors.Is(errWithMsg, target) {
			t.Errorf("WithMessage(%v, %q): is not target", tt.err, tt.message)
		}
	}
}
