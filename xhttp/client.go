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
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zaibyte/pkg/config"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xlog"
	"golang.org/x/net/http2"
)

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
	// DefaultTransport is a h2c transport and backward-compatible with HTTP/1.1.
	DefaultTransport = &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (conn net.Conn, e error) {
			return net.Dial(network, addr)
		},
		DisableCompression: true, // For zai, most req/resp body are binary, so compression maybe useless.
		AllowHTTP:          true,
	}
)

// NewDefaultClientH2C creates a Client with default configs.
func NewDefaultClientH2C() (*Client, error) {

	return NewClient(0, nil), nil
}

// NewDefaultClientH2 creates a TLS Client with default configs.
func NewDefaultClientH2(certFile, keyFile string) (*Client, error) {

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(certBytes) {
		return nil, errors.New("failed to append certs from PEM")
	}

	tc := &tls.Config{
		RootCAs:            cp,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	tp := &http2.Transport{
		TLSClientConfig:    tc,
		DisableCompression: true,
	}

	return NewClient(0, tp), nil
}

// NewClient creates a Client.
// If clientCnt == 0, use defaultClientCnt.
// If transport == nil, use DefaultTransport.
func NewClient(clientCnt int, transport *http2.Transport) *Client {

	config.Adjust(&clientCnt, defaultClientCnt)

	if transport == nil {
		transport = DefaultTransport
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
// A >= 400 status code DO cause an error.
// All >= 400 response will be closed.
//
// On error, any Response can be ignored.
func (c *Client) Request(ctx context.Context,
	method, url, reqID string, body io.Reader,
) (resp *http.Response, err error) {

	url = c.addScheme(url)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}
	if reqID == "" {
		reqID = xlog.NextReqID()
	}
	req.Header.Set(xlog.ReqIDFieldName, reqID)

	hc := c.NextClient()
	resp, err = hc.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode/100 >= 4 {

		err = errors.New(http.StatusText(resp.StatusCode))

		if resp.ContentLength > 0 && method != http.MethodHead {
			buf, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				return resp, err2
			}
			// See ReplyError for more details.
			err = errors.New(string(buf[:len(buf)-1])) // drop \n
		}
		return
	}
	return
}

// --- Default API ---- //
// --- All HTTP Servers in zai will have these APIs ---- //

const defaultTimeout = 3 * time.Second

// Debug opens/closes a server logger's debug level.
func (c *Client) Debug(addr string, on bool, reqID string) (err error) {

	cmd := "off"
	if on {
		cmd = "on"
	}

	url := addr + "/v1/debug-log/" + cmd
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.Request(ctx, http.MethodPut, url, reqID, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return nil
}

// Version returns the code version of a server.
func (c *Client) Version(addr, reqID string) (ver version.Info, err error) {

	url := addr + "/v1/code-version"
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := c.Request(ctx, http.MethodGet, url, reqID, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&ver)
	return
}
