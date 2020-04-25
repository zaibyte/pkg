/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xhttp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/config"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xlog"
	"golang.org/x/net/http2"
)

const UserAgent = "Go-zai-xhttp"

// Client is xhttp client.
type Client struct {
	cs []*http.Client
	id uint64
}

// NextClient uses Round Robin to chose a client.
// For HTTP/2, reuse connections may damage performance if the load is too high,
// so we may need more clients.
func (c *Client) NextClient() *http.Client {
	next := atomic.AddUint64(&c.id, 1) % uint64(len(c.cs))
	return c.cs[next]
}

var (
	defaultClientCnt int = 16 // 16 is enough for most cases.
	// defaultTransport is a h2c transport and backward-compatible with HTTP/1.1.
	defaultTransport = &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (conn net.Conn, e error) {
			return net.Dial(network, addr)
		},
		DisableCompression: true,
		AllowHTTP:          true,
	}
)

// NewDefaultClient creates a Client with default configs.
func NewDefaultClient() *Client {

	return NewClient(0, nil)
}

// NewClient creates a Client.
// If clientCnt == 0, use defaultClientCnt.
func NewClient(clientCnt int, transport http.RoundTripper) *Client {

	config.Adjust(&clientCnt, defaultClientCnt)
	if transport == nil {
		transport = defaultTransport
	}

	cs := make([]*http.Client, clientCnt)
	for i := range cs {
		cs[i] = &http.Client{
			Transport: transport,
		}
	}
	return &Client{
		cs,
		0,
	}
}

const (
	defaultDialTimeout = 3 * time.Second
	// By default, the max object size is 32MB, so 3s is enough.
	defaultRespHeaderTimeout = 3 * time.Second
	defaultIdleConnsPerHost  = 5
)

var defaultH1Transport = newH1Transport(defaultDialTimeout, defaultRespHeaderTimeout, defaultIdleConnsPerHost)

// NewDefaultH1Client creates a http.Client with default HTTP/1.1 Transport.
func NewDefaultH1Client() *Client {

	cs := make([]*http.Client, 1) // No need more than 1 client in HTTP/1.1
	cs[0] = &http.Client{
		Transport: defaultH1Transport,
	}
	return &Client{
		cs,
		0,
	}
}

// NewH1Client creates a http.Client with HTTP/1.1 Transport.
func NewH1Client(dial, resp time.Duration, maxIdle int) *Client {

	cs := make([]*http.Client, 1) // No need more than 1 client in HTTP/1.1
	cs[0] = &http.Client{
		Transport: newH1Transport(dial, resp, maxIdle),
	}
	return &Client{
		cs,
		0,
	}
}

func newH1Transport(dial, resp time.Duration, maxIdle int) *http.Transport {

	return &http.Transport{
		DisableCompression:    true,
		MaxIdleConnsPerHost:   maxIdle,
		ResponseHeaderTimeout: resp,

		DialContext: (&net.Dialer{
			Timeout:   dial,
			KeepAlive: 75 * time.Second,
		}).DialContext,
	}
}

func (c *Client) Put(url, reqID string, body io.Reader, timeout time.Duration) (resp *http.Response, err error) {

	url = addScheme(url)
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return
	}

	return c.Do(req, reqID, timeout)
}

func (c *Client) Get(url, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	url = addScheme(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	return c.Do(req, reqID, timeout)
}

func (c *Client) Delete(url, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	url = addScheme(url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}

	return c.Do(req, reqID, timeout)
}

func addScheme(url string) string {
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	return url
}

// Do sends an HTTP request and returns an HTTP response.
//
// A non-2xx status code DO cause an error.
// All non-2xx response will be closed in zai.
//
// On error, any Response can be ignored.
func (c *Client) Do(req *http.Request, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	if reqID == "" {
		reqID = xlog.NextReqID()
	}
	req.Header.Set(xlog.ReqIDHeader, reqID)
	req.Header.Set("User-Agent", UserAgent)

	if timeout != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		req = req.WithContext(ctx)
	}

	hc := c.NextClient()
	resp, err = hc.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode/100 != 2 { // See ReplyError for more details.
		buf, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			CloseResp(resp)
			return resp, err2
		}
		CloseResp(resp)
		err = errors.New(string(buf[:len(buf)-1])) // drop \n
		return
	}
	return
}

// CloseResp close http.Response gracefully.
func CloseResp(resp *http.Response) {
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

// --- Default Handle API ---- //

// Debug open/close a server logger's debug level.
func (c *Client) Debug(addr string, on bool, reqID string) (err error) {

	cmd := "off"
	if on {
		cmd = "on"
	}
	url := addr + "/debug-log/" + cmd
	url = addScheme(url)
	resp, err := c.Get(url, reqID, time.Second)
	if err != nil {
		return
	}
	defer CloseResp(resp)

	return nil
}

// Version returns the code version of a server.
func (c *Client) Version(addr, reqID string) (ver version.Info, err error) {

	url := addr + "/code-version"
	url = addScheme(url)
	resp, err := c.Get(url, reqID, time.Second)
	if err != nil {
		return
	}
	defer CloseResp(resp)

	err = json.NewDecoder(resp.Body).Decode(&ver)
	return
}

// Ping check a server health.
// Ping also return the server's boxID,
// it makes Keeper response boxID easier.
func (c *Client) Ping(addr, reqID string, timeout time.Duration) (boxID int64, err error) {

	url := addr + "/ping"
	url = addScheme(url)
	resp, err := c.Get(url, reqID, timeout)
	if err != nil {
		return
	}
	defer CloseResp(resp)

	return strconv.ParseInt(resp.Header.Get(xlog.BoxIDHeader), 10, 64)
}
