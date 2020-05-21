/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xhttp

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/zaibyte/pkg/version"

	"github.com/zaibyte/pkg/xlog"
)

var (
	testBoxID  int64
	testLogDir string

	testServer  *httptest.Server
	testSrvAddr string

	testClient *Client
)

func makeTestServer() (err error) {

	appName := "test-xhttp"

	testLogDir, err = ioutil.TempDir(os.TempDir(), appName)
	if err != nil {
		return
	}

	lc := &xlog.ServerConfig{
		AccessLogOutput: filepath.Join(testLogDir, "access.log"),
		ErrorLogOutput:  filepath.Join(testLogDir, "error.log"),
		ErrorLogLevel:   "info",
		Rotate:          xlog.RotateConfig{},
	}

	testBoxID = 1

	al, err := lc.MakeLogger("test-xhttp", testBoxID)
	if err != nil {
		return
	}

	srv := NewServer(&ServerConfig{
		AppName: appName,
	}, al)
	srv.srv.Handler = srv.toH2CHandler()
	testServer = httptest.NewServer(srv.srv.Handler)
	testSrvAddr = testServer.URL

	testClient = NewClient(1, defaultTransport)

	return nil
}

func cleanup() {
	testServer.Close()
	os.RemoveAll(testLogDir)
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

func TestServerH2cH1(t *testing.T) {

	resp, err := testClient.Request(context.Background(), http.MethodHead, testSrvAddr+"/v1/ping", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseResp(resp)

	if resp.ProtoMajor != 2 {
		t.Fatal("proto should be 2 with HTTP/2 client")
	}

	c1 := http.DefaultClient // HTTP/1.1 should be ok too.
	resp, err = c1.Head(testSrvAddr + "/v1/ping")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseResp(resp)

	if resp.ProtoMajor != 1 {
		t.Fatal("proto should be 1 with HTTP/1.1 client")
	}
}

func TestServerPing(t *testing.T) {

	boxID, err := testClient.Ping(testSrvAddr, "", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != testBoxID {
		t.Fatal("boxID mismatch", boxID)
	}

	c1 := http.DefaultClient // HTTP/1.1 should be ok too.

	resp, err := c1.Head(testSrvAddr + "/v1/ping")
	if err != nil {
		t.Fatal(err)
	}

	boxID, err = strconv.ParseInt(resp.Header.Get(xlog.BoxIDField), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if boxID != testBoxID {
		t.Fatal("boxID mismatch")
	}
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

func TestMustHaveHeader(t *testing.T) {

	resp, err := testClient.Request(context.Background(), http.MethodHead, testSrvAddr+"/v1/ping", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer CloseResp(resp)

	if resp.Header.Get(xlog.BoxIDField) != strconv.Itoa(int(testBoxID)) {
		t.Fatal("boxID mismatch")
	}

	if resp.Header.Get(xlog.ReqIDField) == "" {
		t.Fatal("request ID missing")
	}

	_, _, err = xlog.ParseReqID(resp.Header.Get(xlog.ReqIDField))
	if err != nil {
		t.Fatal("illegal request ID")
	}
}
