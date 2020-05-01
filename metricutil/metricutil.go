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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zaibyte/pkg/config"
	"github.com/zaibyte/pkg/typeutil"
	iid "github.com/zaibyte/pkg/uid/instanceid"
	"github.com/zaibyte/pkg/xlog"
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
func Push(cfg *Config, instanceID string) {

	if len(cfg.PushAddress) == 0 || cfg.PushJob == "" {
		xlog.Info("disable Prometheus push client")
		return
	}

	config.Adjust(&instanceID, iid.Get())

	if cfg.PushInterval.Duration == zeroDuration {
		cfg.PushInterval.Duration = defaultPushInterval
	}

	xlog.Info("start Prometheus push client")

	go prometheusPushClient(cfg, instanceID)
}

// prometheusPushClient pushes metrics to Prometheus Pushgateway.
func prometheusPushClient(cfg *Config, instanceID string) {
	pusher := push.New(cfg.PushAddress, cfg.PushJob).
		Gatherer(prometheus.DefaultGatherer).
		Grouping("instance", instanceID)

	for {
		err := pusher.Push()
		if err != nil {
			xlog.Error(fmt.Sprintf("could not push metrics to Prometheus Pushgateway: %s", err.Error()))
		}

		time.Sleep(cfg.PushInterval.Duration)
	}
}
