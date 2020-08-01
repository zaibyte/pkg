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

package xdigest

import (
	"fmt"
	"hash/crc32"
	"reflect"
	"testing"
)

type b struct {
	b []byte
}

func TestWriteNil(t *testing.T) {
	b := new(b)
	h := crc32.New(CrcTbl)
	h.Write(b.b)
	fmt.Println(h.Sum32())
}

type Router struct {
	bitmap [32]byte
}

func TestBitmap(t *testing.T) {
	r := new(Router)
	r.setNopResp(33)
	fmt.Println(r.isNopResp(33))
	for i := 0; i < 256; i++ {
		if i == 33 {
			if !r.isNopResp(uint8(i)) {
				t.Fatal("should true")
			}

		} else {
			if r.isNopResp(uint8(i)) {
				t.Fatal("should false")
			}
		}
	}
	fmt.Println(r.isNopResp(32))
}

func (r *Router) setNopResp(method uint8) {
	byteOff := method / 8
	bitOff := method % 8
	r.bitmap[byteOff] |= 1 << bitOff
}

func (r *Router) isNopResp(method uint8) bool {
	byteOff := method / 8
	bitOff := method % 8
	return r.bitmap[byteOff]&(1<<bitOff) != 0
}

type HandlerFunc func(reqid uint64, req []byte, extraReq []byte) (resp []byte, err error)

type AA struct {
}

func (a *AA) shit(reqid uint64, req []byte, extraReq []byte) (resp []byte, err error) {
	return
}

func TestReqType(t *testing.T) {
	var hf HandlerFunc
	a := new(AA)
	hf = a.shit
	printType(hf)

}

func printType(hf HandlerFunc) {
	fv := reflect.Indirect(reflect.ValueOf(hf))
	ft := fv.Type()
	reqT := ft.In(0)
	if reqT == reflect.TypeOf(uint64(1)) {
		fmt.Println("uint")
	}
	if reqT == reflect.TypeOf([]byte{}) {
		fmt.Println("byte")
	}
}
