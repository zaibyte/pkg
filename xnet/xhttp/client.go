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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zaibyte/pkg/uid"
	"github.com/zaibyte/pkg/version"
	"github.com/zaibyte/pkg/xnet"
)

// Client is an xhttp client.
type Client struct {
	c         *http.Client
	encrypted bool
	addScheme func(url string) string
}

const (
	defaultDialTimeout         = 3000 * time.Millisecond
	defaultRespTimeout         = 16 * time.Second
	defaultKeepAlive           = 75 * time.Second
	defaultMaxIdleConns        = 100
	defaultMaxIdleConnsPerHost = 10
	defaultIdleConnTimeout     = 90 * time.Second
)

var (
	// DefaultTransport is a HTTP/1.1 transport.
	DefaultTransport = &http.Transport{
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		IdleConnTimeout:       defaultIdleConnTimeout,
		ResponseHeaderTimeout: defaultRespTimeout,
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
	}
)

// NewDefaultClient creates a Client with default configs.
func NewDefaultClient() (*Client, error) {
	return NewClient(false, "", "")
}

// NewClient creates a Client with tls configs.
func NewClient(encrypted bool, certFile, keyFile string) (*Client, error) {

	tp := DefaultTransport

	if certFile == "" || keyFile == "" {
		encrypted = false
	}

	if encrypted != false && certFile != "" && keyFile != "" {
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

		tp.TLSClientConfig = tc
	}

	return NewClientWithTransport(tp), nil
}

// NewClientWithTransport creates a Client with a transport.
// If transport == nil, use DefaultTransport.
func NewClientWithTransport(transport *http.Transport) *Client {

	if transport == nil {
		transport = DefaultTransport
	}

	addScheme := addHTTPScheme
	if transport.TLSClientConfig != nil {
		addScheme = addHTTPSScheme
	}

	return &Client{
		c: &http.Client{
			Transport: transport,
		},
		encrypted: transport.TLSClientConfig != nil,
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
func (c *Client) Request(ctx context.Context, method, url, reqID string, buf []byte) (resp *http.Response, err error) {

	url = c.addScheme(url)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(buf))
	if err != nil {
		return
	}
	if reqID == "" {
		reqID = strconv.FormatUint(uid.MakeReqID(), 10)
	}
	req.Header.Set(ReqIDHeader, reqID)

	if !c.encrypted {
		req.Header.Set(ChecksumHeader, strconv.Itoa(int(xnet.Checksum(buf))))
	}

	resp, err = c.c.Do(req)
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
