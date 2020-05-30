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

// Package xlog provides logger features.
//
// All log entries are encoded in JSON,
// and time format is ISO8601 ("2006-01-02T15:04:05.000Z0700")
package xlog

import (
	"path/filepath"
	"strings"

	"github.com/zaibyte/pkg/config/settings"

	"github.com/zaibyte/pkg/config"
	"go.uber.org/zap"
)

// TimeFormat is used for parsing log entry's time field.
const (
	ISO8601TimeFormat = "2006-01-02T15:04:05.000Z0700"
)

// RotateConfig is partly copy from logro's Config,
// hiding details in logro.
type RotateConfig struct {
	// Maximum size of a log file before it gets rotated.
	// Unit is MB.
	MaxSize int64 `toml:"max_size"`
	// Maximum number of backup log files to retain.
	MaxBackups int
	// Timestamp in backup log file. Default(false) is UTC time.
	LocalTime bool `toml:"local_time"`
}

// ServerConfig is the log configs of a http server application.
type ServerConfig struct {
	AccessLogOutput string       `toml:"access_log_output"`
	ErrorLogOutput  string       `toml:"error_log_output"`
	ErrorLogLevel   string       `toml:"error_log_level"`
	Rotate          RotateConfig `toml:"rotate"`
}

// MakeAppLogger init global error logger and returns loggers for application.
func (c *ServerConfig) MakeAppLogger(appName string, boxID int64) (el *ErrorLogger, al *AccessLogger, err error) {

	config.Adjust(&c.ErrorLogOutput, filepath.Join(settings.DefaultLogRoot, appName, "error.log"))
	config.Adjust(&c.ErrorLogLevel, "info")

	el, err = NewErrorLogger(c.ErrorLogOutput, c.ErrorLogLevel, &c.Rotate)
	if err != nil {
		return
	}

	config.Adjust(&c.AccessLogOutput, filepath.Join(settings.DefaultLogRoot, appName, "access.log"))
	al, err = NewAccessLogger(c.AccessLogOutput, &c.Rotate)
	if err != nil {
		return
	}

	InitGlobalLogger(el, boxID)

	return
}

// Types here for hiding zap, don't need to know zap outside.
//
// String constructs a field with the given key and value.
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Int constructs a field with the given key and value.
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 constructs a field with the given key and value.
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64 constructs a field with the given key and value.
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// ReqID constructs a field with the Request ID key and value.
func ReqID(reqID string) zap.Field {
	return String(strings.ToLower(ReqIDFieldName), reqID)
}

// BoxID constructs a field with the Box ID key and value.
func BoxID(boxID int64) zap.Field {
	return Int64(strings.ToLower(BoxIDFieldName), boxID)
}
