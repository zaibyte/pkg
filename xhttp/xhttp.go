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

// Package xhttp provides http server & client methods.
//
// xhttp supports HTTP/2 & HTTP/1.1 (if it's in h2c and the request protocol is HTTP/1).
package xhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/zaibyte/pkg/xlog"
)

// Reply replies HTTP request, return the written bytes length & status code,
// we need the status code for access log.
//
// Usage:
// As return function in http Handler.
//
// Warn:
// Be sure you have called xlog.InitGlobalLogger.
// If any wrong in the write resp process, it would be written into the log.

// ReplyCode replies to the request with the empty message and HTTP code.
func ReplyCode(w http.ResponseWriter, statusCode int) (written, status int) {

	return ReplyJson(w, nil, statusCode)
}

// ReplyError replies to the request with the specified error message and HTTP code.
func ReplyError(w http.ResponseWriter, msg string, statusCode int) (written, status int) {

	if msg == "" {
		msg = http.StatusText(statusCode)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	written, err := fmt.Fprintln(w, msg)
	if err != nil {
		xlog.WarnWithReqID(makeReplyErrMsg(err), w.Header().Get(xlog.ReqIDField))
	}
	return written, statusCode
}

// ReplyJson replies to the request with specified ret(in JSON) and HTTP code.
func ReplyJson(w http.ResponseWriter, ret interface{}, statusCode int) (written, status int) {

	var msg []byte
	if ret != nil {
		msg, _ = json.Marshal(ret)
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(msg)))
	w.WriteHeader(statusCode)
	written, err := w.Write(msg)
	if err != nil {
		xlog.WarnWithReqID(makeReplyErrMsg(err), w.Header().Get(xlog.ReqIDField))
	}
	return written, statusCode
}

// ReplyBin replies to the request with specified ret(in Binary) and length.
func ReplyBin(w http.ResponseWriter, ret io.Reader, length int64) (written, status int) {

	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	n, err := io.CopyN(w, ret, length)
	if err != nil {
		xlog.WarnWithReqID(makeReplyErrMsg(err), w.Header().Get(xlog.ReqIDField))
	}
	return int(n), http.StatusOK
}

func makeReplyErrMsg(err error) string {
	return fmt.Sprintf("write resp failed: %s", err.Error())
}

// FillPath fills the julienschmidt/httprouter style path.
func FillPath(path string, kv map[string]string) string {
	if kv == nil {
		return path
	}

	for k, v := range kv {
		path = strings.Replace(path, ":"+k, v, 1)
	}
	return path
}
