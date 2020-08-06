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

package xtcp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/templexxx/tsc"

	"github.com/templexxx/xhex"
	"github.com/zaibyte/pkg/xstrconv"

	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/xdigest"

	_ "github.com/zaibyte/pkg/xlog/xlogtest"
	"github.com/zaibyte/pkg/xrpc"
)

// TODO create a map to keep put obj
func testPutFunc(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
	return nil
}

func testGetFunc(reqid uint64, oid [16]byte) (objData xrpc.Byteser, err error) {
	return
}

func testDeleteFunc(reqid uint64, oid [16]byte) error {
	return nil
}

func getRandomAddr() string {
	rand.Seed(tsc.UnixNano())
	return fmt.Sprintf("127.0.0.1:%d", rand.Intn(20000)+10000)
}

func TestRequestTimeout(t *testing.T) {

	addr := getRandomAddr()

	s := NewServer(addr, nil, func(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}, testGetFunc, testDeleteFunc)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	objData := make([]byte, 16)
	rand.Read(objData)
	digest := xdigest.Sum32(objData)
	_, oid := uid.MakeOID(1, 1, digest, 16, uid.NormalObj)

	for i := 0; i < 10; i++ {
		err := c.PutObj(0, oid, objData, time.Millisecond)
		if err == nil {
			t.Fatalf("Timeout error must be returned")
		}
		if err != xrpc.ErrTimeout {
			t.Fatalf("Unexpected error returned: [%s]", err)
		}
	}
}

func TestClient_GetObj(t *testing.T) {
	addr := getRandomAddr()

	stor := make(map[[16]byte][]byte)

	s := NewServer(addr, nil, func(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
		o := make([]byte, len(objData.Bytes()))
		copy(o, objData.Bytes())
		stor[oid] = o
		return nil
	}, func(reqid uint64, oid [16]byte) (objData xrpc.Byteser, err error) {
		_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
		objData = xrpc.GetNBytes(int(size))
		o := stor[oid]
		objData.Write(o)
		return
	}, testDeleteFunc)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	req := make([]byte, xrpc.MaxBytesSizeInPool*2)
	rand.Read(req)

	for i := 0; i < 7; i++ {

		size := (1 << i) * 1024
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		_, oid := uid.MakeOID(1, 1, digest, uint32(size), uid.NormalObj)

		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
	}

	for oid, objBytes := range stor {
		b := make([]byte, 32)
		xhex.Encode(b, oid[:])
		bf, err := c.GetObj(0, xstrconv.ToString(b), 0)
		if err != nil {
			t.Fatal(err)
		}
		_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
		act := make([]byte, size)
		n, err := bf.Read(act)
		if err != nil {
			bf.Close()
			t.Fatal(err, size, n)
		}
		if !bytes.Equal(act, objBytes) {
			bf.Close()
			t.Fatal("obj data mismatch")
		}
		bf.Close()
	}
}

func TestClient_DeleteObj(t *testing.T) {
	addr := getRandomAddr()

	stor := make(map[[16]byte][]byte)

	s := NewServer(addr, nil,
		func(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
			o := make([]byte, len(objData.Bytes()))
			copy(o, objData.Bytes())
			stor[oid] = o
			return nil
		},
		func(reqid uint64, oid [16]byte) (objData xrpc.Byteser, err error) {

			o, ok := stor[oid]
			if !ok {
				return nil, xrpc.ErrNotFound
			}
			_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
			objData = xrpc.GetNBytes(int(size))
			objData.Write(o)
			return
		},
		func(reqid uint64, oid [16]byte) error {
			_, ok := stor[oid]
			if !ok {
				return xrpc.ErrNotFound
			}
			delete(stor, oid)
			return nil
		})
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	req := make([]byte, xrpc.MaxBytesSizeInPool*2)
	rand.Read(req)

	for i := 0; i < 7; i++ {

		size := (1 << i) * 1024
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		_, oid := uid.MakeOID(1, 1, digest, uint32(size), uid.NormalObj)

		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
	}

	deleted := make([][16]byte, 3)
	cnt := 0
	for oid := range stor {
		if cnt >= 3 {
			break
		}
		b := make([]byte, 32)
		xhex.Encode(b, oid[:])
		err := c.DeleteObj(0, xstrconv.ToString(b), 0)
		if err != nil {
			t.Fatal(err)
		}
		var do [16]byte
		copy(do[:], oid[:])
		deleted[cnt] = do
		cnt++
	}

	for oid, objBytes := range stor {
		b := make([]byte, 32)
		xhex.Encode(b, oid[:])
		bf, err := c.GetObj(0, xstrconv.ToString(b), 0)
		if err != nil {
			t.Fatal(err)
		}
		_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
		act := make([]byte, size)
		n, err := bf.Read(act)
		if err != nil {
			bf.Close()
			t.Fatal(err, size, n)
		}
		if !bytes.Equal(act, objBytes) {
			bf.Close()
			t.Fatal("obj data mismatch")
		}
		bf.Close()
	}

	for _, oid := range deleted {
		b := make([]byte, 32)
		xhex.Encode(b, oid[:])
		bf, err := c.GetObj(0, xstrconv.ToString(b), 0)

		assert.Nil(t, bf)
		assert.Equal(t, xrpc.ErrNotFound, err)
	}
}

func TestClient_GetObj_Concurrency(t *testing.T) {
	addr := getRandomAddr()

	stor := new(sync.Map)

	s := NewServer(addr, nil,
		func(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
			_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
			o := make([]byte, size)
			n, err := objData.Read(o)
			if err != nil {
				return xrpc.ErrInternalServer
			}
			if n != int(size) {
				return xrpc.ErrInternalServer
			}
			stor.Store(oid, o)
			return nil
		}, func(reqid uint64, oid [16]byte) (objData xrpc.Byteser, err error) {
			_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
			objData = xrpc.GetNBytes(int(size))
			v, ok := stor.Load(oid)
			if !ok {
				return nil, xrpc.ErrNotFound
			}
			o := v.([]byte)
			objData.Write(o)
			return
		}, testDeleteFunc)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, nil)
	c.Start()
	defer c.Stop()

	req := make([]byte, 1024*1024)
	rand.Read(req)

	oids := make([]string, 18)

	for i := 0; i < 18; i++ {

		size := (1 << i) * 2
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		_, oid := uid.MakeOID(1, 1, digest, uint32(size), uid.NormalObj)
		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids {
		wg.Add(1)
		go func(oid string) {
			defer wg.Done()
			bf, err := c.GetObj(0, oid, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer bf.Close()

			_, _, _, _, size, _, _ := uid.ParseOID(oid)
			act := make([]byte, size)
			bf.Read(act)
			var ob [16]byte // Using byte array to save function stack space.
			xhex.Decode(ob[:16], xstrconv.ToBytes(oid))
			v, ok := stor.Load(ob)
			if !ok {
				t.Fatal("not found")
			}
			if !bytes.Equal(act, v.([]byte)) {
				t.Fatal("get obj data mismatch")
			}

		}(oid)
	}

	wg.Wait()
}

func TestClient_GetObj_ConcurrencyTLS(t *testing.T) {
	addr := getRandomAddr()

	certFile := "./ssl-cert-snakeoil.pem"
	keyFile := "./ssl-cert-snakeoil.key"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatalf("cannot load TLS certificates: [%s]", err)
	}
	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	clientCfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	stor := new(sync.Map)

	s := NewServer(addr, serverCfg,
		func(reqid uint64, oid [16]byte, objData xrpc.Byteser) error {
			_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
			o := make([]byte, size)
			n, err := objData.Read(o)
			if err != nil {
				return xrpc.ErrInternalServer
			}
			if n != int(size) {
				return xrpc.ErrInternalServer
			}
			stor.Store(oid, o)
			return nil
		}, func(reqid uint64, oid [16]byte) (objData xrpc.Byteser, err error) {
			_, _, _, _, size, _ := uid.ParseOIDBytes(oid[:])
			objData = xrpc.GetNBytes(int(size))
			v, ok := stor.Load(oid)
			if !ok {
				return nil, xrpc.ErrNotFound
			}
			o := v.([]byte)
			objData.Write(o)
			return
		}, testDeleteFunc)
	if err := s.Start(); err != nil {
		t.Fatalf("cannot start server: %s", err)
	}
	defer s.Stop()

	c := NewClient(addr, clientCfg)
	c.Start()
	defer c.Stop()

	req := make([]byte, 1024*1024)
	rand.Read(req)

	oids := make([]string, 18)

	for i := 0; i < 18; i++ {

		size := (1 << i) * 2
		objData := req[:size]
		digest := xdigest.Sum32(objData)
		_, oid := uid.MakeOID(1, 1, digest, uint32(size), uid.NormalObj)
		err := c.PutObj(0, oid, objData, 0)
		if err != nil {
			t.Fatal(err, size)
		}
		oids[i] = oid
	}

	var wg sync.WaitGroup
	for _, oid := range oids {
		wg.Add(1)
		go func(oid string) {
			defer wg.Done()
			bf, err := c.GetObj(0, oid, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer bf.Close()

			_, _, _, _, size, _, _ := uid.ParseOID(oid)
			act := make([]byte, size)
			bf.Read(act)
			var ob [16]byte // Using byte array to save function stack space.
			xhex.Decode(ob[:16], xstrconv.ToBytes(oid))
			v, ok := stor.Load(ob)
			if !ok {
				t.Fatal("not found")
			}
			if !bytes.Equal(act, v.([]byte)) {
				t.Fatal("get obj data mismatch")
			}

		}(oid)
	}

	wg.Wait()
}
