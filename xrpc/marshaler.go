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

// Marshaler is the interface for types that can be Marshaled & Unmarshaled.
type Marshaler interface {
	// MarshalTo encodes obj into buf and returns the encoded bytes.
	// If the buffer is too small, MarshalTo will panic.
	MarshalTo(buf []byte) ([]byte, error)
	// MarshalLen returns the serial byte size.
	MarshalLen() (int, error)
	// UnmarshalBinary decodes data.
	UnmarshalBinary(data []byte) error
}

// NopMarshaler is a marshaler which has no data.
type NopMarshaler struct{}

func (m *NopMarshaler) MarshalTo(_ []byte) ([]byte, error) {
	return nil, nil
}

func (m *NopMarshaler) MarshalLen() (int, error) {
	return 0, nil
}

func (m *NopMarshaler) UnmarshalBinary(_ []byte) error {
	return nil
}

func (m *NopMarshaler) Free() {}

// MarshalFreer is the interface for type which implements Marshaler
// and supports Free() to put it back to sync.Pool.
type MarshalFreer interface {
	Marshaler
	Free()
}
