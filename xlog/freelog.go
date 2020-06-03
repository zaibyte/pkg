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

// FreeLogger is a logger without time, level, msg ... (any default fields)
// for highly specialised.
type FreeLogger struct {
	l        *zap.Logger
	rotation *logro.Rotation
}

func freeEncoderConf() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,
		LineEnding: zapcore.DefaultLineEnding,
	}
}

// NewFreeLogger return a new FreeLogger.
func NewFreeLogger(outputPath string, rCfg *RotateConfig, fields []zap.Field) (logger *FreeLogger, err error) {

	syncer, err := logro.New(&logro.Config{
		OutputPath: outputPath,
		MaxSize:    rCfg.MaxSize,
		MaxBackups: rCfg.MaxBackups,
		LocalTime:  rCfg.LocalTime,
	})
	if err != nil {
		return
	}

	lvl := zap.NewAtomicLevelAt(zap.InfoLevel)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(freeEncoderConf()), syncer, lvl)
	if fields != nil {
		core = core.With(fields)
	}

	return &FreeLogger{zap.New(core), syncer}, nil
}

// Info logs in info level.
func (l *FreeLogger) Info(f ...zap.Field) {
	l.l.Info("", f...)
}

// Sync syncs FreeLogger.
func (l *FreeLogger) Sync() error {
	return l.l.Sync()
}

// Close closes FreeLogger.
func (l *FreeLogger) Close() error {
	return l.rotation.Close()
}
