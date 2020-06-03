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
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/zaibyte/pkg/xmath"
)

// AccessLogger is used for recording the HTTP server access logs.
type AccessLogger struct {
	fl *FreeLogger
}

// NewAccessLogger creates an AceesLogger.
func NewAccessLogger(outputPath string, rCfg *RotateConfig) (logger *AccessLogger, err error) {
	fl, err := NewFreeLogger(outputPath, rCfg, nil)
	if err != nil {
		return
	}

	return &AccessLogger{fl}, nil
}

// These field names will be added in access log.
// And they are HTTP headers in zai too.
const (
	ReqIDFieldName = "x-zai-request-id"
)

// AccessLogFields shows access logger output fields.
type AccessLogFields struct {
	API           string  `json:"api"`
	Status        int     `json:"status"`
	BodyBytesSent int     `json:"body_bytes_sent"`
	BodyBytesRecv int64   `json:"body_bytes_recv"`
	RequestTime   float64 `json:"request_time"`
	Time          string  `json:"time"`
	ReqID         string  `json:"x-zai-request-id "`
}

// |      name          |  type  |             detail              |		e.g		              |
// |--------------------|--------|---------------------------------|------------------------------|
// | api                | string | see ps 5                        | zai.get                      |
// | status             | int    | HTTP status code                | 200                          |
// | body_bytes_sent    | int    | response body length(written)   | 1                            |
// | body_bytes_recv    | int64  | request body length             | 1                            |
// | request_time       | float64| see ps 2                        | 1.00                         |
// | time               | string | log entry written time(ISO8601) | 2018-12-26T01:09:22.852+0800 |
// | x-zai-request-id   | string |                                 | 100AAOCNvxRDC6AV             |
//
// ps:
// 1.body_bytes_sent
// Not the value of Content-Length in resp header,
// it's the real bytes written in resp.
// It may return a error when writing to resp sometimes,
// so the Content-Length will mislead us.
//
// 2.request_time
// Request processing time in milliseconds(keep two decimals);
// time elapsed between the first bytes were read from the client
// and the log write after the last bytes were sent to the client
//
// 3. time
// Present time in ISO8601TimeFormat,
// (error logger use this fmt to)
//
// 4. api
// handle's name, is used to distinguish different requests for
// analysing logs in the future

// Write writes entries to AccessLogger.
func (l *AccessLogger) Write(apiName string, r *http.Request, start time.Time, reqID string, written, status int) {

	now := time.Now()
	l.fl.Info(
		zap.String("api", apiName),
		zap.Int("status", status),
		zap.Int64("body_bytes_recv", r.ContentLength),
		zap.Int("body_bytes_sent", written),
		zap.Float64("request_time", xmath.Round(now.Sub(start).Seconds()*1000, 2)),
		zap.String(ReqIDFieldName, reqID),
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
