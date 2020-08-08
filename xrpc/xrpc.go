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

	"github.com/zaibyte/pkg/xbytes"
)

// Objecter is the object RPC client.
type Objecter interface {
	// Start Objecter.
	Start() error
	// Stop Objecter, release resource.
	Stop() error
	// Put puts object to the ZBuf node which Objecter connected.
	PutObj(reqid uint64, oid string, objData []byte, timeout time.Duration) error
	// Get gets object from the ZBuf node which Objecter connected.
	GetObj(reqid uint64, oid string, timeout time.Duration) (obj io.ReadCloser, err error)
	// Delete deletes object in the ZBuf node which Objecter connected.
	DeleteObj(reqid uint64, oid string, timeout time.Duration) error
}

type PutFunc func(reqid uint64, oid [16]byte, objData xbytes.Buffer) error

type GetFunc func(reqid uint64, oid [16]byte) (objData xbytes.Buffer, err error)

type DeleteFunc func(reqid uint64, oid [16]byte) error
