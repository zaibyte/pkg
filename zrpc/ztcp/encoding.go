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
//
// The MIT License (MIT)
//
// Copyright (c) 2014 Aliaksandr Valialkin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This file contains code derived from gorpc.
// The main logic & codes are copied from gorpc.

package ztcp

import (
	"bufio"
	"compress/flate"
	"encoding/gob"
	"io"
)

// RegisterType registers the given type to send via rpc.
//
// The client must register all the response types the server may send.
// The server must register all the request types the client may send.
//
// There is no need in registering base Go types such as int, string, bool,
// float64, etc. or arrays, slices and maps containing base Go types.
//
// There is no need in registering argument and return value types
// for functions and methods registered via Dispatcher.
func RegisterType(x interface{}) {
	gob.Register(x)
}

type wireRequest struct {
	ID      uint64
	Request interface{}
}

type wireResponse struct {
	ID       uint64
	Response interface{}
	Error    string
}

type messageEncoder struct {
	e  *gob.Encoder
	bw *bufio.Writer
	zw *flate.Writer
	ww *bufio.Writer
}

func (e *messageEncoder) Close() error {
	if e.zw != nil {
		return e.zw.Close()
	}
	return nil
}

func (e *messageEncoder) Flush() error {
	if e.zw != nil {
		if err := e.ww.Flush(); err != nil {
			return err
		}
		if err := e.zw.Flush(); err != nil {
			return err
		}
	}
	if err := e.bw.Flush(); err != nil {
		return err
	}
	return nil
}

func (e *messageEncoder) Encode(msg interface{}) error {
	return e.e.Encode(msg)
}

func newMessageEncoder(w io.Writer, bufferSize int) *messageEncoder {
	bw := bufio.NewWriterSize(w, bufferSize)

	ww := bw
	var zw *flate.Writer

	return &messageEncoder{
		e:  gob.NewEncoder(ww),
		bw: bw,
		zw: zw,
		ww: ww,
	}
}

type messageDecoder struct {
	d  *gob.Decoder
	zr io.ReadCloser
}

func (d *messageDecoder) Close() error {
	if d.zr != nil {
		return d.zr.Close()
	}
	return nil
}

func (d *messageDecoder) Decode(msg interface{}) error {
	return d.d.Decode(msg)
}

// TODO no need compress.
func newMessageDecoder(r io.Reader, bufferSize int) *messageDecoder {
	br := bufio.NewReaderSize(r, bufferSize)

	rr := br
	var zr io.ReadCloser

	return &messageDecoder{
		d:  gob.NewDecoder(rr),
		zr: zr,
	}
}
