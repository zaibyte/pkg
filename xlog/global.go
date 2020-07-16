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
	_global.Error("", string(p))
	return len(p), nil
}

func Error(msg string) {
	_global.Error("", msg)
}

func Info(msg string) {
	_global.Info("", msg)
}

func Warn(msg string) {
	_global.Warn("", msg)
}

func Debug(msg string) {
	_global.Debug("", msg)
}

func Fatal(msg string) {
	_global.Fatal("", msg)
}

func Panic(msg string) {
	_global.Panic("", msg)
}

func Errorf(format string, args ...interface{}) {
	_global.Errorf("", format, args)
}

func Infof(format string, args ...interface{}) {
	_global.Infof("", format, args)
}

func Warnf(format string, args ...interface{}) {
	_global.Warnf("", format, args)
}

func Debugf(format string, args ...interface{}) {
	_global.Debugf("", format, args)
}

func Fatalf(format string, args ...interface{}) {
	_global.Fatalf("", format, args)
}

func Panicf(format string, args ...interface{}) {
	_global.Panicf("", format, args)
}

func ErrorID(reqid, msg string) {
	_global.Error(reqid, msg)
}

func InfoID(reqid, msg string) {
	_global.Info(reqid, msg)
}

func WarnID(reqid, msg string) {
	_global.Warn(reqid, msg)
}

func DebugID(reqid, msg string) {
	_global.Debug(reqid, msg)
}

func FatalID(reqid, msg string) {
	_global.Fatal(reqid, msg)
}

func PanicID(reqid, msg string) {
	_global.Panic(reqid, msg)
}

func ErrorIDf(reqid, format string, args ...interface{}) {
	_global.Errorf(reqid, format, args)
}

func InfoIDf(reqid, format string, args ...interface{}) {
	_global.Infof(reqid, format, args)
}

func WarnIDf(reqid, format string, args ...interface{}) {
	_global.Warnf(reqid, format, args)
}

func DebugIDf(reqid, format string, args ...interface{}) {
	_global.Debugf(reqid, format, args)
}

func FatalIDf(reqid, format string, args ...interface{}) {
	_global.Fatalf(reqid, format, args)
}

func PanicIDf(reqid, format string, args ...interface{}) {
	_global.Panicf(reqid, format, args)
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
