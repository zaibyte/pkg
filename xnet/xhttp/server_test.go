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

package xhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xlog"
	"github.com/zaibyte/pkg/xlog/xlogtest"
)

var (
	testBoxID  int64
	testLogDir string

	testServer  *httptest.Server
	testSrvAddr string

	testClient *Client
)

func makeTestServer() (err error) {

	xlogtest.New()

	srv := NewServer(&ServerConfig{
		Encrypted:         false,
		CertFile:          "",
		KeyFile:           "",
		IdleTimeout:       0,
		ReadHeaderTimeout: 0,
	})
	testServer = httptest.NewServer(srv.srv.Handler)
	testSrvAddr = testServer.URL

	testClient, _ = NewDefaultClient()

	return nil
}

func cleanup() {
	testServer.Close()
	xlogtest.Close()
}

func TestMain(m *testing.M) {
	err := makeTestServer()
	if err != nil {
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestServerDebug(t *testing.T) {

	err := testClient.Debug(testSrvAddr, true, "")
	if err != nil {
		t.Fatal(err)
	}

	if xlog.GetLvl() != "debug" {
		t.Fatal("debug on failed")
	}

	err = testClient.Debug(testSrvAddr, false, "")
	if err != nil {
		t.Fatal(err)
	}

	if xlog.GetLvl() != "info" {
		t.Fatal("debug off failed")
	}
}

func TestServerLimit(t *testing.T) {

	errMsg := make(chan string, 10)
	wg := &sync.WaitGroup{}
	wg.Add(10) // Although the limit is 1, but the op is too fast, so we may need more to pass the test.
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			_, err := testClient.Version(testSrvAddr, "")
			if err != nil {
				errMsg <- err.Error()
			}
		}()
	}
	wg.Wait()
	close(errMsg)
	eCnt := 0
	for msg := range errMsg {
		eCnt++
		if msg != http.StatusText(http.StatusTooManyRequests) {
			t.Fatal(fmt.Sprintf("unexpect error msg, exp: %s, act:%s", http.StatusText(http.StatusTooManyRequests), msg))
		}
	}
	if eCnt <= 0 || eCnt > 9 {
		t.Fatal("unexpect error count", eCnt)
	}
}

func TestServerVersion(t *testing.T) {

	ret, err := testClient.Version(testSrvAddr, "")
	if err != nil {
		t.Fatal(err)
	}

	if ret.Version != version.ReleaseVersion {
		t.Fatal("ReleaseVersion mismatch")
	}
	if ret.GitBranch != version.GitBranch {
		t.Fatal("GitBranch mismatch")
	}
	if ret.GitHash != version.GitHash {
		t.Fatal("GitBranch mismatch")
	}
}

func TestFillPath(t *testing.T) {
	path := "/test/:k0/:k1/:k2"
	kv := make(map[string]string)
	kv["k0"] = "v0"
	kv["k1"] = "v1"
	kv["k2"] = "v2"

	act := FillPath(path, kv)
	exp := "/test/v0/v1/v2"
	if act != exp {
		t.Fatal("mismatch")
	}
}
