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

import "go.uber.org/zap"

var (
	_global *ErrorLogger
)

// Init Global var.
// warn: It's unsafe for concurrent use.
func InitGlobalLogger(logger *ErrorLogger) {
	_global = logger
}

// Write implements io.Writer
func Write(p []byte) (n int, err error) {
	_global.Error(string(p))
	return len(p), nil
}

func Error(msg string, f ...zap.Field) {
	_global.Error(msg, f...)
}

func Info(msg string, f ...zap.Field) {
	_global.Info(msg, f...)
}

func Warn(msg string, f ...zap.Field) {
	_global.Warn(msg, f...)
}

func Debug(msg string, f ...zap.Field) {
	_global.Debug(msg, f...)
}

func Fatal(msg string, f ...zap.Field) {
	_global.Fatal(msg, f...)
}

func Panic(msg string, f ...zap.Field) {
	_global.Panic(msg, f...)
}

func ErrorWithReqID(msg, reqID string) {
	_global.Error(msg, ReqID(reqID))
}

func InfoWithReqID(msg, reqID string) {
	_global.Info(msg, ReqID(reqID))
}

func WarnWithReqID(msg, reqID string) {
	_global.Warn(msg, ReqID(reqID))
}

func DebugWithReqID(msg, reqID string) {
	_global.Debug(msg, ReqID(reqID))
}

func FatalWithReqID(msg, reqID string) {
	_global.Fatal(msg, ReqID(reqID))
}

func PanicWithReqID(msg, reqID string) {
	_global.Panic(msg, ReqID(reqID))
}

// Sync syncs _global.
func Sync() error {
	return _global.Sync()
}

// Close closes _global.
func Close() error {
	return _global.Close()
}

func SetLevel(level string) error {

	return _global.SetLevel(level)
}

// GetLvl returns lvl in string.
func GetLvl() string {
	return _global.GetLvl()
}

// GetLogger returns _global logger.
func GetLogger() *ErrorLogger {
	return _global
}
