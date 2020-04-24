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
	"strings"

	"go.uber.org/zap"
)

var (
	_global *ErrorLogger
	_boxID  int64 // xlog only need string.
)

// Init Global var.
// I separate logger & boxID here,
// because for non-keeper components, they won't know the boxID
// until get boxID from Keeper, so there is a gap.

// InitGlobalLogger init the _global.
// warn: It's unsafe for concurrent use.
func InitGlobalLogger(logger *ErrorLogger) {
	_global = logger
}

// InitGlobalLogger init the _global.
// warn: It's unsafe for concurrent use.
func InitBoxID(boxID int64) {
	_boxID = boxID
}

// GetBoxID returns global boxID.
func GetBoxID() int64 {
	return _boxID
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

func makeReqIDField(reqID string) zap.Field {
	return String(strings.ToLower(ReqIDHeader), reqID)
}

func makeBoxIDField() zap.Field {
	return Int64(strings.ToLower(BoxIDHeader), _boxID)
}

func ErrorWithReqID(msg, reqID string) {

	_global.Error(msg, makeReqIDField(reqID), makeBoxIDField())
}

func InfoWithReqID(msg, reqID string) {

	_global.Info(msg, makeReqIDField(reqID), makeBoxIDField())
}

func WarnWithReqID(msg, reqID string) {

	_global.Warn(msg, makeReqIDField(reqID), makeBoxIDField())
}

func DebugWithReqID(msg, reqID string) {

	_global.Debug(msg, makeReqIDField(reqID), makeBoxIDField())
}

func FatalWithReqID(msg, reqID string) {

	_global.Fatal(msg, makeReqIDField(reqID), makeBoxIDField())
}

func PanicWithReqID(msg, reqID string) {

	_global.Panic(msg, makeReqIDField(reqID), makeBoxIDField())
}

// Sync syncs _global.
func Sync() error {
	return _global.Sync()
}

// Close closes _global.
func Close() error {
	return _global.Close()
}

// DebugOn enables debug level.
func DebugOn() {
	_global.DebugOn()
}

// DebugOff enables info level.
func DebugOff() {
	_global.DebugOff()
}

// GetLvl returns lvl in string.
func GetLvl() string {
	return _global.GetLvl()
}

// GetLogger returns _global logger.
func GetLogger() *ErrorLogger {
	return _global
}
