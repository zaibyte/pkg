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
	"errors"
	"fmt"
	"log"
)

type ExampleDispatcherService struct {
	state int
}

func (s *ExampleDispatcherService) Get() int { return s.state }

func (s *ExampleDispatcherService) Set(x int) { s.state = x }

func (s *ExampleDispatcherService) GetError42() (int, error) {
	if s.state == 42 {
		return 0, errors.New("error42")
	}
	return s.state, nil
}

func (s *ExampleDispatcherService) privateFunc(string) { s.state = 0 }

func ExampleDispatcher_serviceCalls() {
	d := NewDispatcher()

	service := &ExampleDispatcherService{
		state: 123,
	}

	// Register exported service functions
	d.AddService("MyService", service)

	// Start rpc server serving registered service.
	addr := "127.0.0.1:7892"
	s := NewTCPServer(addr, d.NewHandlerFunc())
	if err := s.Start(); err != nil {
		log.Fatalf("Cannot start rpc server: [%s]", err)
	}
	defer s.Stop()

	// Start rpc client connected to the server.
	c := NewTCPClient(addr)
	c.Start()
	defer c.Stop()

	// Create client wrapper for calling service functions.
	dc := d.NewServiceClient("MyService", c)

	res, err := dc.Call("Get", nil)
	fmt.Printf("Get=%+v, %+v\n", res, err)

	service.state = 456
	res, err = dc.Call("Get", nil)
	fmt.Printf("Get=%+v, %+v\n", res, err)

	res, err = dc.Call("Set", 78)
	fmt.Printf("Set=%+v, %+v, %+v\n", res, err, service.state)

	res, err = dc.Call("GetError42", nil)
	fmt.Printf("GetError42=%+v, %+v\n", res, err)

	service.state = 42
	res, err = dc.Call("GetError42", nil)
	fmt.Printf("GetError42=%+v, %+v\n", res, err)

	// Output:
	// Get=123, <nil>
	// Get=456, <nil>
	// Set=<nil>, <nil>, 78
	// GetError42=78, <nil>
	// GetError42=<nil>, error42
}
