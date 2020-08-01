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

package xtcp

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/zaibyte/pkg/xlog"

	"github.com/zaibyte/pkg/xrpc"
)

// Router helps constructing HandlerFunc for routing requests according request method.
type Router struct {
	handlers []*HandlerFunc
	cnt      int

	nopResps   [32]byte
	bytesResps [32]byte
}

// NewRouter creates a router.
func NewRouter() *Router {
	return &Router{
		handlers: make([]*HandlerFunc, maxMethod+1),
	}
}

// AddFunc adds HandlerFunc to router.
//
func (r *Router) AddFunc(method uint8, hf HandlerFunc) error {

	if method == 0 {
		return xrpc.ErrInvalidMethod
	}

	if hf == nil {
		return errors.New("illegal HandlerFunc: nil")
	}

	if r.handlers[method] != nil {
		return errors.New(fmt.Sprintf("method: %d has been already registered", method))
	}

	r.handlers[method] = &hf

	fv := reflect.Indirect(reflect.ValueOf(hf))
	ft := fv.Type()
	respT := ft.Out(0)
	if respT == reflect.TypeOf(&xrpc.NopMarshaler{}) {
		r.setNopResp(method)
	}
	if respT == reflect.TypeOf(&xrpc.Buffer{}) {
		r.setBytesResp(method)
	}

	r.cnt++
	return nil
}

func (r *Router) setNopResp(method uint8) {
	byteOff := method / 8
	bitOff := method % 8
	r.nopResps[byteOff] |= 1 << bitOff
}

func (r *Router) isNopResp(method uint8) bool {
	byteOff := method / 8
	bitOff := method % 8
	return r.nopResps[byteOff]&(1<<bitOff) == 1
}

func (r *Router) setBytesResp(method uint8) {
	byteOff := method / 8
	bitOff := method % 8
	r.bytesResps[byteOff] |= 1 << bitOff
}

func (r *Router) isBytesResp(method uint8) bool {
	byteOff := method / 8
	bitOff := method % 8
	return r.bytesResps[byteOff]&(1<<bitOff) == 1
}

// AddToClient adds router info to Client.
func (r *Router) AddToClient(c *Client) {

	c.router = r
}

// AddToServer adds router to Server.
func (r *Router) AddToServer(s *Server) {

	s.Router = r
}

// Handle handles requests.
func (r *Router) Handle(reqid uint64, method uint8, req, extraReq *xrpc.Buffer) (resp *xrpc.Buffer, err error) {
	if r.cnt == 0 {
		xlog.Panic("register at least one HandlerFunc")
	}

	if method == 0 {
		return nil, xrpc.ErrInvalidMethod
	}

	h := r.handlers[method]
	if h == nil {
		return nil, xrpc.ErrNotImplemented
	}

	hf := *h

	return hf(reqid, req, extraReq)
}
