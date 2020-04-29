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
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zaibyte/pkg/typeutil"
	"github.com/zaibyte/pkg/xlog"
)

const zeroDuration = time.Duration(0)

type Config struct {
	InstanceName string            `json:"instance_name" toml:"instance_name"`
	PushJob      string            `json:"push_job" toml:"push_job"`
	PushAddress  string            `json:"push_address" toml:"push_address"`
	PushInterval typeutil.Duration `json:"push_interval" toml:"push_interval"`
}

const (
	defaultInstanceName = "unknown"
	defaultPushInterval = 15 * time.Second
)

// Push metrics in background.
func Push(cfg *Config) {

	if cfg.InstanceName == "" {
		cfg.InstanceName = makeInstanceName()
	}

	if cfg.PushInterval.Duration == zeroDuration {
		cfg.PushInterval.Duration = defaultPushInterval
	}

	if len(cfg.PushAddress) == 0 {
		xlog.Info("disable Prometheus push client")
		return
	}

	xlog.Info("start Prometheus push client")

	go prometheusPushClient(cfg)
}

// prometheusPushClient pushs metrics to Prometheus Pushgateway.
func prometheusPushClient(cfg *Config) {
	pusher := push.New(cfg.PushAddress, cfg.PushJob).
		Gatherer(prometheus.DefaultGatherer).
		Grouping("instance", cfg.InstanceName)

	for {
		err := pusher.Push()
		if err != nil {
			xlog.Error(fmt.Sprintf("could not push metrics to Prometheus Pushgateway: %s", err.Error()))
		}

		time.Sleep(cfg.PushInterval.Duration)
	}
}

func makeInstanceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return defaultInstanceName
	}
	return hostname
}
