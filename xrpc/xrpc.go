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

import (
	"io"
	"time"
)

type PutObjReq struct {
	GroupID uint16
	TS      uint32
	Digest  uint32
	Size    uint32
	Otype   uint8
}

type GetObjReq struct {
	GroupID uint16
	TS      uint32
	Digest  uint32
}

type DelObjReq struct {
	GroupID uint16
	TS      uint32
	Digest  uint32
}

type Client interface {
	Start() error
	Stop() error
	PutObj(reqid uint64, req *PutObjReq, obj []byte, timeout time.Duration) error
	GetObj(reqid uint64, req *GetObjReq, timeout time.Duration) (obj io.ReadCloser, err error)
	DelObj(reqid uint64, req *DelObjReq, timeout time.Duration) error
}

type Server interface {
	Start() error
	Stop() error
}
