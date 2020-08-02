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

import "fmt"

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
	_global.Error(0, string(p))
	return len(p), nil
}

func Error(msg string) {
	_global.Error(0, msg)
}

func Info(msg string) {
	_global.Info(0, msg)
}

func Warn(msg string) {
	_global.Warn(0, msg)
}

func Debug(msg string) {
	_global.Debug(0, msg)
}

func Fatal(msg string) {
	_global.Fatal(0, msg)
}

func Panic(msg string) {
	_global.Panic(0, msg)
}

func Errorf(format string, args ...interface{}) {
	_global.Error(0, fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	_global.Info(0, fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	_global.Warn(0, fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	_global.Debug(0, fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	_global.Fatal(0, fmt.Sprintf(format, args...))
}

func Panicf(format string, args ...interface{}) {
	_global.Panic(0, fmt.Sprintf(format, args...))
}

func ErrorID(reqid uint64, msg string) {
	_global.Error(reqid, msg)
}

func InfoID(reqid uint64, msg string) {
	_global.Info(reqid, msg)
}

func WarnID(reqid uint64, msg string) {
	_global.Warn(reqid, msg)
}

func DebugID(reqid uint64, msg string) {
	_global.Debug(reqid, msg)
}

func FatalID(reqid uint64, msg string) {
	_global.Fatal(reqid, msg)
}

func PanicID(reqid uint64, msg string) {
	_global.Panic(reqid, msg)
}

func ErrorIDf(reqid uint64, format string, args ...interface{}) {
	_global.Error(reqid, fmt.Sprintf(format, args...))
}

func InfoIDf(reqid uint64, format string, args ...interface{}) {
	_global.Info(reqid, fmt.Sprintf(format, args...))
}

func WarnIDf(reqid uint64, format string, args ...interface{}) {
	_global.Warn(reqid, fmt.Sprintf(format, args...))
}

func DebugIDf(reqid uint64, format string, args ...interface{}) {
	_global.Debug(reqid, fmt.Sprintf(format, args...))
}

func FatalIDf(reqid uint64, format string, args ...interface{}) {
	_global.Fatal(reqid, fmt.Sprintf(format, args...))
}

func PanicIDf(reqid uint64, format string, args ...interface{}) {
	_global.Panic(reqid, fmt.Sprintf(format, args...))
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
