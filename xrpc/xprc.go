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
// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file contains code derived from Dragonboat.
// The main logic & codes are copied from Dragonboat.

// Package xrpc contains structs, interfaces and function definitions required
// to build customized Zai RPC modules.
//
// Structs, interfaces and functions defined in the xrpc package are only
// required when building your customized RPC modules. You can safely skip this
// package if you plan to use the default built-in Zai RPC modules provided by Zai.
//
// Structs, interfaces and functions defined in the xrpc package are not
// considered as a part of Zai's public APIs. Breaking changes might
// happen in the coming minor releases.
package xrpc

import (
	"context"

	pb "github.com/lni/dragonboat/v3/raftpb"
	colf "github.com/zaibyte/pkg/colf"
)

type ObjPutHandler func(reqid uint64, req colf.PutObjReq) error

type ObjGetHandler func(reqid uint64, req colf.GetObjReq) ([]byte, error)

type ObjDelHandler func(reqid uint64, req colf.DelObjReq) error

// RequestHandler is the handler function type for handling received message
// batch. Received message batches should be passed to the request handler to
// have them processed by Dragonboat.
type RequestHandler func(req pb.MessageBatch)

// IChunkHandler is the handler interface to handle incoming snapshot chunks.
type IChunkHandler interface {
	// AddChunk adds a new snapshot chunk to the snapshot chunk sink. All chunks
	// belong to the snapshot will be combined into the snapshot image and then
	// be passed to Dragonboat once all member chunks are received.
	AddChunk(chunk pb.Chunk) bool
}

// IChunkSink is the interface of snapshot chunk sink. IChunkSink is used to
// accept received snapshot chunks.
type IChunkSink interface {
	IChunkHandler
	// Close closes the sink instance and releases all resources held by it.
	Close()
	// Tick moves forward the internal logic clock. It is suppose to be called
	// roughly every second.
	Tick()
}

// IConnection is the interface used by the Raft RPC module for sending Raft
// messages. Each IConnection works for a specified target nodehost instance,
// it is possible for a target to have multiple concurrent IConnection
// instances.
type IConnection interface {
	// Close closes the IConnection instance.
	Close()
	// SendMessageBatch sends the specified message batch to the target. It is
	// recommended to deliver the message batch to the target in order to enjoy
	// best possible performance, but out of order delivery is allowed at the
	// cost of reduced performance.
	SendMessageBatch(batch pb.MessageBatch) error
}

// ISnapshotConnection is the interface used by the Raft RPC module for sending
// snapshot chunks. Each ISnapshotConnection works for a specified target
// nodehost instance.
type ISnapshotConnection interface {
	// Close closes the ISnapshotConnection instance.
	Close()
	// SendChunk sends the snapshot chunk to the target. It is
	// recommended to have the snapshot chunk delivered in order for the best
	// performance, but out of order delivery is allowed at the cost of reduced
	// performance.
	SendChunk(chunk pb.Chunk) error
}

// IRaftRPC is the interface to be implemented by a customized Raft RPC
// module. A Raft RPC module is responsible for exchanging Raft messages
// including snapshot chunks between nodehost instances.
type IRaftRPC interface {
	// Name returns the type name of the IRaftRPC instance.
	Name() string
	// Start launches the Raft RPC module and make it ready to start sending and
	// receiving Raft messages.
	Start() error
	// Stop stops the Raft RPC instance.
	Stop()
	// GetConnection returns an IConnection instance responsible for
	// sending Raft messages to the specified target nodehost.
	GetConnection(ctx context.Context, target string) (IConnection, error)
	// GetSnapshotConnection returns an ISnapshotConnection instance used for
	// sending snapshot chunks to the specified target nodehost.
	GetSnapshotConnection(ctx context.Context,
		target string) (ISnapshotConnection, error)
}
