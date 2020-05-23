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
	"github.com/templexxx/logro"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorLogger is used for recording the common application log,
// xlog also provides global logger for more convenient.
// In practice, ErrorLogger is just a global logger's container,
// it won't be used directly.
type ErrorLogger struct {
	l        *zap.Logger
	lvl      zap.AtomicLevel
	rotation *logro.Rotation
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
	r, err := logro.New(&logro.Config{
		OutputPath: outputPath,
		MaxSize:    rCfg.MaxSize,
		MaxBackups: rCfg.MaxBackups,
		LocalTime:  rCfg.LocalTime,
	})
	if err != nil {
		return
	}

	lvl := zap.NewAtomicLevel()
	err = lvl.UnmarshalText([]byte(level))
	if err != nil {
		return
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(defaultEncoderConf()), r, lvl)

	return &ErrorLogger{
		l:        zap.New(core),
		rotation: r,
	}, nil
}

// default without caller and stack trace,
func defaultEncoderConf() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

// Write implements io.Writer
func (l *ErrorLogger) Write(p []byte) (n int, err error) {
	l.Error(string(p))
	return len(p), nil
}

func (l *ErrorLogger) Error(msg string, f ...zap.Field) {
	l.l.Error(msg, f...)
}

func (l *ErrorLogger) Info(msg string, f ...zap.Field) {
	l.l.Info(msg, f...)
}

func (l *ErrorLogger) Warn(msg string, f ...zap.Field) {
	l.l.Warn(msg, f...)
}

func (l *ErrorLogger) Debug(msg string, f ...zap.Field) {
	l.l.Debug(msg, f...)
}

func (l *ErrorLogger) Fatal(msg string, f ...zap.Field) {
	l.l.Fatal(msg, f...)
}

func (l *ErrorLogger) Panic(msg string, f ...zap.Field) {
	l.l.Panic(msg, f...)
}

// Sync syncs ErrorLogger.
func (l *ErrorLogger) Sync() error {
	return l.l.Sync()
}

func (l *ErrorLogger) Close() error {
	return l.rotation.Close()
}

// DebugOn enable debug level.
func (l *ErrorLogger) DebugOn() {
	l.lvl.SetLevel(zap.DebugLevel)
}

// DebugOff enable info level.
func (l *ErrorLogger) DebugOff() {
	l.lvl.SetLevel(zap.InfoLevel)
}

// GetLvl return lvl in string.
func (l *ErrorLogger) GetLvl() string {
	return l.lvl.String()
}
