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

package ztcp

import (
	"errors"
	"fmt"
)

// Router helps constructing Handler for routing requests according request method.
type Router struct {
	handlers []*Handler
	cnt      int
}

// NewRouter creates a router.
func NewRouter() *Router {
	return &Router{
		handlers: make([]*Handler, maxMethod+1),
	}
}

// Marshaler is the interface for types that can be Marshaled.
type Marshaler interface {
	MarshalTo([]byte) (int, error)
	Size() int
}

// Response is the interface for ztcp response.
type Response interface {
	Marshaler
	error
}

// Handler is a server handler function.
//
// reqBody and resp types is Marshaler.
//
// Hint: use Router for HandlerFunc construction.
type Handler func(reqid uint64, method uint8, reqBody Marshaler) (resp Response)

func (r *Router) AddFunc(method uint8, handler *Handler) error {

	if handler == nil {
		return errors.New("illegal Handler: nil")
	}

	if r.handlers[method] != nil {
		return errors.New(fmt.Sprintf("ztcp.Router, method: %d has been already registered", method))
	}

	r.cnt++
	return nil
}

// NewHandler returns Handler serving all the functions registered via AddFunc().
//
// The returned Handler must be assigned to Server.Handler or
// passed to New*Server().
func (r *Router) NewHandler() Handler {
	if r.cnt == 0 {
		logPanic("ztcp.Router: register at least one Handler")
	}

	hs := copyRouter(r.handlers)

	return func(reqid uint64, method uint8, reqBody Marshaler) (resp Response) {

		return routerRequest(hs, reqid, method, reqBody)
	}
}

func copyRouter(hs []*Handler) []*Handler {
	ret := make([]*Handler, maxMethod+1)
	for i := range hs {
		ret[i] = hs[i]
	}
	return ret
}

type nopResponse struct {
	errMsg string
}

func (r *nopResponse) MarshalTo([]byte) (int, error) {
	return 0, nil
}

func (r *nopResponse) Size() int {
	return 0
}

func (r *nopResponse) Error() string {
	return r.errMsg
}

func routerRequest(hs []*Handler, reqid uint64, method uint8, reqBody Marshaler) (resp Response) {

	h := hs[method]
	if h == nil {
		return &nopResponse{errMsg: "method not found"}
	}

	hf := *h
	return hf(reqid, method, reqBody)
}
