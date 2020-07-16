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
	"github.com/zaibyte/nanozap"
	"github.com/zaibyte/nanozap/zapcore"
	"github.com/zaibyte/nanozap/zaproll"
)

// RotateConfig is partly copy from zaproll's Config,
// hiding details in zaproll.
type RotateConfig struct {
	// Maximum size of a log file before it gets rotated.
	// Unit is MB.
	MaxSize int64 `toml:"max_size"`
	// Maximum number of backup log files to retain.
	MaxBackups int
	// Timestamp in backup log file. Default(false) is UTC time.
	LocalTime bool `toml:"local_time"`
}

// ErrorLogger is used for recording the common application log,
// xlog also provides global logger for more convenient.
// In practice, ErrorLogger is just a global logger's container,
// it won't be used directly.
type ErrorLogger struct {
	l        *nanozap.Logger
	lvl      nanozap.AtomicLevel
	rotation *zaproll.Rotation
}

// ErrLogFields shows error logger output fields.
//
// Warn:
// Sometimes, there is no "x-zai-request-id"" or "x-zai-box-id".
// (It's not from any request)
type ErrLogFields struct {
	Level string `json:"level"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	ReqID string `json:"x-zai-request-id"`
	BoxID int64  `json:"x-zai-box-id"`
}

// NewErrorLogger returns a logger with its properties.
//
// Legal Levels:
// info: "info", "INFO", ""
// debug: "debug", "DEBUG"
// warn: "warn", "WARN"
// error: "error", "ERROR"
// panic: "panic", "PANIC"
// fatal: "fatal", "FATAL"
func NewErrorLogger(outputPath, level string, rCfg *RotateConfig) (logger *ErrorLogger, err error) {

	r, err := zaproll.New(&zaproll.Config{
		OutputPath: outputPath,
		MaxSize:    rCfg.MaxSize,
		MaxBackups: rCfg.MaxBackups,
		LocalTime:  rCfg.LocalTime,
	})
	if err != nil {
		return
	}

	lvl := nanozap.NewAtomicLevel()
	err = lvl.UnmarshalText([]byte(level))
	if err != nil {
		return
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(defaultEncoderConf()), r, lvl)

	return &ErrorLogger{
		l:        nanozap.New(core),
		rotation: r,
	}, nil
}

// default without caller and stack trace,
func defaultEncoderConf() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		ReqIDKey:       "reqid",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochMillisTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

// Write implements io.Writer
func (l *ErrorLogger) Write(p []byte) (n int, err error) {
	l.Error("", string(p))
	return len(p), nil
}

func (l *ErrorLogger) Error(reqid, msg string) {
	l.l.Error(reqid, msg)
}

func (l *ErrorLogger) Info(reqid, msg string) {
	l.l.Info(reqid, msg)
}

func (l *ErrorLogger) Warn(reqid, msg string) {
	l.l.Warn(reqid, msg)
}

func (l *ErrorLogger) Debug(reqid, msg string) {
	l.l.Debug(reqid, msg)
}

func (l *ErrorLogger) Fatal(reqid, msg string) {
	l.l.Fatal(reqid, msg)
}

func (l *ErrorLogger) Panic(reqid, msg string) {
	l.l.Panic(reqid, msg)
}

func (l *ErrorLogger) Errorf(reqid, format string, args ...interface{}) {
	l.l.Errorf(reqid, format, args)
}

func (l *ErrorLogger) Infof(reqid, format string, args ...interface{}) {
	l.l.Infof(reqid, format, args)
}

func (l *ErrorLogger) Warnf(reqid, format string, args ...interface{}) {
	l.l.Warnf(reqid, format, args)
}

func (l *ErrorLogger) Debugf(reqid, format string, args ...interface{}) {
	l.l.Debugf(reqid, format, args)
}

func (l *ErrorLogger) Fatalf(reqid, format string, args ...interface{}) {
	l.l.Fatalf(reqid, format, args)
}

func (l *ErrorLogger) Panicf(reqid, format string, args ...interface{}) {
	l.l.Panicf(reqid, format, args)
}

// Sync syncs ErrorLogger.
func (l *ErrorLogger) Sync() error {
	return l.l.Sync()
}

func (l *ErrorLogger) Close() error {
	l.l.Close()
	return l.rotation.Close()
}

func (l *ErrorLogger) SetLevel(level string) error {
	lvl := nanozap.NewAtomicLevel()
	err := lvl.UnmarshalText([]byte(level))
	if err != nil {
		return err
	}

	l.lvl.SetLevel(lvl.Level())
	return nil
}

// GetLvl return lvl in string.
func (l *ErrorLogger) GetLvl() string {
	return l.lvl.String()
}
