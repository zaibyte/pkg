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
	cs        []*http.Client
	id        uint64
	addScheme func(url string) string
}

// NextClient uses Round Robin to chose a client.
// For HTTP/2, reuse connections may damage performance if the load is too high,
// so we may need more clients.
func (c *Client) NextClient() *http.Client {
	next := atomic.AddUint64(&c.id, 1) % uint64(len(c.cs))
	return c.cs[next]
}

const (
	defaultClientCnt int = 16 // 16 is enough for most cases.
)

var (
	// defaultTransport is a h2c transport and backward-compatible with HTTP/1.1.
	defaultTransport = &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (conn net.Conn, e error) {
			return net.Dial(network, addr)
		},
		DisableCompression: true, // For zai, most objects are binary, so compression maybe useless.
		AllowHTTP:          true,
	}
)

// NewDefaultClient creates a Client with default configs.
func NewDefaultClient() *Client {

	return NewClient(0, nil)
}

// NewClient creates a Client.
// If clientCnt == 0, use defaultClientCnt.
// If transport == nil, use defaultTransport.
func NewClient(clientCnt int, transport *http2.Transport) *Client {

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

	addScheme := addHTTPScheme
	if transport.TLSClientConfig != nil {
		addScheme = addHTTPSScheme
	}

	return &Client{
		cs:        cs,
		id:        0,
		addScheme: addScheme,
	}
}

// Head sends an HTTP PUT request and returns an HTTP response.
func (c *Client) Put(url, reqID string, body io.Reader, timeout time.Duration) (resp *http.Response, err error) {

	return c.Do(url, http.MethodPut, reqID, body, timeout)
}

// Head sends an HTTP GET request and returns an HTTP response.
func (c *Client) Get(url, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	return c.Do(url, http.MethodGet, reqID, nil, timeout)
}

// Head sends an HTTP DELETE request and returns an HTTP response.
func (c *Client) Delete(url, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	return c.Do(url, http.MethodDelete, reqID, nil, timeout)
}

// Head sends an HTTP HEAD request and returns an HTTP response.
func (c *Client) Head(url, reqID string, timeout time.Duration) (resp *http.Response, err error) {

	return c.Do(url, http.MethodHead, reqID, nil, timeout)
}

func addHTTPSScheme(url string) string {
	return addScheme(url, "https://")
}

func addHTTPScheme(url string) string {
	return addScheme(url, "http://")
}

// addScheme adds HTTP scheme if need.
func addScheme(url string, scheme string) string {
	if !strings.HasPrefix(url, scheme) {
		url = scheme + url
	}
	return url
}

// Do sends an HTTP request and returns an HTTP response.
//
// A non-2xx status code DO cause an error.
// All non-2xx response will be closed.
//
// On error, any Response can be ignored.
func (c *Client) Do(url, method, reqID string, body io.Reader, timeout time.Duration) (resp *http.Response, err error) {

	url = c.addScheme(url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
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
// --- All HTTP Servers in zai will have these APIs ---- //

// Debug opens/closes a server logger's debug level.
func (c *Client) Debug(addr string, on bool, reqID string) (err error) {

	cmd := "off"
	if on {
		cmd = "on"
	}
	url := addr + "/debug-log/" + cmd
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
	resp, err := c.Get(url, reqID, time.Second)
	if err != nil {
		return
	}
	defer CloseResp(resp)

	err = json.NewDecoder(resp.Body).Decode(&ver)
	return
}

// Ping checks a server health and returns the server's boxID,
func (c *Client) Ping(addr, reqID string, timeout time.Duration) (boxID int64, err error) {

	url := addr + "/ping"
	resp, err := c.Head(url, reqID, timeout)
	if err != nil {
		return
	}
	defer CloseResp(resp)

	return strconv.ParseInt(resp.Header.Get(xlog.BoxIDHeader), 10, 64)
}
