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

package xlog

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// AccessLogger is used for recording the server access logs.
type AccessLogger struct {
	fl *FreeLogger
}

// NewAccessLogger creates an AceesLogger.
func NewAccessLogger(outputPath string, rCfg *RotateConfig) (logger *AccessLogger, err error) {
	fl, err := NewFreeLogger(outputPath, rCfg)
	if err != nil {
		return
	}
	return &AccessLogger{fl}, nil
}

// These field names will be added in access log.
// And they are HTTP headers in zai too.
const (
	ReqIDFieldName = "X-zai-Request-ID"
	BoxIDFieldName = "X-zai-Box-ID"
)

// AccessLogFields shows access logger output fields.
type AccessLogFields struct {
	API           string  `json:"api"`
	RemoteAddr    string  `json:"remote_addr"`
	Status        int     `json:"status"`
	BodyBytesSent int     `json:"body_bytes_sent"`
	BodyBytesRecv int64   `json:"body_bytes_recv"`
	RequestTime   float64 `json:"request_time"`
	Time          string  `json:"time"`
	ReqID         string  `json:"x-zai-request-id "`
	BoxID         int64   `json:"x-zai-box-id"`
}

// |      name          |  type  |             detail              |		e.g		              |
// |--------------------|--------|---------------------------------|------------------------------|
// | api                | string | see ps 5                        | zai.get                      |
// | remote_addr        | string |                                 | 192.168.1.3                  |
// | status             | int    | HTTP status code                | 200                          |
// | body_bytes_sent    | int    | response body length(written)   | 1                            |
// | body_bytes_recv    | int64  | request body length             | 1                            |
// | request_time       | float64| see ps 2                        | 1.00                         |
// | time               | string | log entry written time(ISO8601) | 2018-12-26T01:09:22.852+0800 |
// | x-zai-request-id   | string |                                 | 100AAOCNvxRDC6AV             |
// | x-zai-box-id       | int64  |                                 | 1                            |
//
// ps:
// 1.body_bytes_sent
// Not the value of Content-Length in resp header,
// it's the real bytes written in resp.
// It may return a error when writing to resp sometimes,
// so the Content-Length will mislead us
//
// 2.request_time
// Request processing time in seconds with a milliseconds resolution;
// time elapsed between the first bytes were read from the client
// and the log write after the last bytes were sent to the client
//
// 3. remote_addr
// The request source.
// If there is a proxy in front of server, the remote_addr may be wrong
// so set header X-Real-IP = $remote_addr in your proxy
//
// 4. time
// Present time in ISO8601TimeFormat,
// (error logger use this fmt to)
//
// 5. api
// handle's name, is used to distinguish different requests for
// analysing logs in the future

// Write writes entry to AccessLogger.
func (l *AccessLogger) Write(apiName string, r *http.Request,
	start time.Time, reqID string, written, status int) {

	remoteAddr := r.Header.Get("X-Real-IP")
	if remoteAddr == "" {
		remoteAddr = strings.Split(r.RemoteAddr, ":")[0]
	}

	now := time.Now()
	l.fl.Write(
		String("api", apiName),
		String("remote_addr", remoteAddr),
		String("time", now.Format(ISO8601TimeFormat)),
		Int("status", status),
		Int64("body_bytes_recv", r.ContentLength),
		Int("body_bytes_sent", written),
		Float64("request_time", round(now.Sub(start).Seconds()*1000, 2)),
		String(strings.ToLower(ReqIDFieldName), reqID),
		Int64(strings.ToLower(BoxIDFieldName), _boxID),
	)
}

// Sync syncs AccessLogger.
func (l *AccessLogger) Sync() (err error) {
	return l.fl.Sync()
}

// Close closes AccessLogger.
func (l *AccessLogger) Close() (err error) {
	return l.fl.Close()
}

func round(f float64, n int) float64 {
	pow10n := math.Pow10(n)
	return math.Trunc(f*pow10n+0.5) / pow10n
}

// default max_pid = num_processors * 1024,
// or max_pid = 32768 when num_processors < 32.
// Uint16 may not enough, so uint32.
var _pid = uint32(os.Getpid())

// NextReqID returns a request ID.
// warn: maybe not unique but it's acceptable.
func NextReqID() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], _pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

// ParseReqID gets pid & time from a reqID.
func ParseReqID(reqID string) (pid uint32, t time.Time, err error) {

	b, err := base64.URLEncoding.DecodeString(reqID)
	if err != nil {
		return
	}

	pid = binary.LittleEndian.Uint32(b[:4])
	nt := int64(binary.LittleEndian.Uint64(b[4:]))
	t = time.Unix(0, nt)
	return
}
