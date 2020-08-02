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

	"github.com/zaibyte/pkg/xlog"

	"github.com/zaibyte/pkg/xrpc"
)

// Router helps constructing HandlerFunc for routing requests according request method.
type Router struct {
	// 256 is the max method numbers.
	handlers []*HandlerFunc // Using a slice as "map", 256 * 8bytes = 1KB cache friendly.
	cnt      int            // Indicates valid methods.
	methods  *bitmap        // Every bit is the flag shows has the method or not. 1 means has, otherwise not.
}

// NewRouter creates a router.
func NewRouter() *Router {
	return &Router{
		handlers: make([]*HandlerFunc, maxMethod+1),
		methods:  &bitmap{p: make([]byte, 32)},
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
	r.methods.set(method)
	r.cnt++

	return nil
}

// AddToClient adds methods bitmap to Client.
func (r *Router) AddToClient(c *Client) {

	c.methods = r.methods
}

// AddToServer adds router to Server.
func (r *Router) AddToServer(s *Server) {

	s.Router = r
}

// Handle handles requests.
func (r *Router) Handle(reqid uint64, method uint8, req *xrpc.Buffer, mreqLen int) (resp xrpc.MarshalFreer, err error) {
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

	return hf(reqid, req, mreqLen)
}

type bitmap struct {
	p []byte
}

func (b *bitmap) set(method uint8) {
	byteOff := method / 8
	bitOff := method % 8
	b.p[byteOff] |= 1 << bitOff
}

func (b *bitmap) has(method uint8) bool {
	byteOff := method / 8
	bitOff := method % 8
	return !(b.p[byteOff]&(1<<bitOff) == 0)
}
