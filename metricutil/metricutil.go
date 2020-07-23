/*
 * Copyright (c) 2020. Temple3x (temple3x@gmail.com)
 * Copyright 2016 PingCAP, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package metricutil provides functions to push metrics to Prometheus Pushgateway.
package metricutil

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zaibyte/pkg/typeutil"
	"github.com/zaibyte/pkg/zlog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const zeroDuration = time.Duration(0)

type Config struct {
	PushJob      string            `json:"push_job" toml:"push_job"`
	PushAddress  string            `json:"push_address" toml:"push_address"`
	PushInterval typeutil.Duration `json:"push_interval" toml:"push_interval"`
}

const (
	defaultPushInterval = 15 * time.Second
)

// Push metrics in background.
func Push(cfg *Config, boxID uint32, instanceID string) {

	if len(cfg.PushAddress) == 0 || cfg.PushJob == "" {
		zlog.Info("disable Prometheus push client")
		return
	}

	if boxID == 0 {
		panic("boxID must not be 0")
	}

	if instanceID == "" {
		panic("instanceID must not be empty")
	}

	if cfg.PushInterval.Duration == zeroDuration {
		cfg.PushInterval.Duration = defaultPushInterval
	}

	zlog.Info("start Prometheus push client")

	go prometheusPushClient(cfg, boxID, instanceID)
}

// prometheusPushClient pushes metrics to Prometheus Pushgateway.
func prometheusPushClient(cfg *Config, boxID uint32, instanceID string) {
	pusher := push.New(cfg.PushAddress, cfg.PushJob).
		Gatherer(prometheus.DefaultGatherer).
		Grouping("box", strconv.Itoa(int(boxID))).
		Grouping("instance", instanceID)

	for {
		err := pusher.Push()
		if err != nil {
			zlog.Error(fmt.Sprintf("could not push metrics to Prometheus Pushgateway: %s", err.Error()))
		}

		time.Sleep(cfg.PushInterval.Duration)
	}
}
