// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/prometheus-community/ipmi_exporter/freeipmi"
)

const (
	namespace   = "ipmi"
	targetLocal = ""
)

type collector interface {
	Name() CollectorName
	Cmd() string
	Args() []string
	Collect(output freeipmi.Result, ch chan<- prometheus.Metric, target ipmiTarget) (int, error)
}

type metaCollector struct {
	target string
	module string
	config *SafeConfig
}

type ipmiTarget struct {
	host   string
	config IPMIConfig
}

var (
	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"'1' if a scrape of the IPMI device was successful, '0' otherwise.",
		[]string{"collector"},
		nil,
	)

	durationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape_duration", "seconds"),
		"Returns how long the scrape took to complete in seconds.",
		nil,
		nil,
	)
)

// Describe implements Prometheus.Collector.
func (c metaCollector) Describe(ch chan<- *prometheus.Desc) {
	// all metrics are described ad-hoc
}

func markCollectorUp(ch chan<- prometheus.Metric, name string, up int) {
	ch <- prometheus.MustNewConstMetric(
		upDesc,
		prometheus.GaugeValue,
		float64(up),
		name,
	)
}

// Collect implements Prometheus.Collector.
func (c metaCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		log.Debugf("Scrape of target %s took %f seconds.", targetName(c.target), duration)
		ch <- prometheus.MustNewConstMetric(
			durationDesc,
			prometheus.GaugeValue,
			duration,
		)
	}()

	config := c.config.ConfigForTarget(c.target, c.module)
	target := ipmiTarget{
		host:   c.target,
		config: config,
	}

	for _, collector := range config.GetCollectors() {
		var up int
		log.Debugf("Running collector: %s", collector.Name())

		fqcmd := path.Join(*executablesPath, collector.Cmd())
		args := collector.Args()
		cfg := config.GetFreeipmiConfig()

		result := freeipmi.Execute(fqcmd, args, cfg, target.host, log.Base())

		up, _ = collector.Collect(result, ch, target)
		markCollectorUp(ch, string(collector.Name()), up)
	}
}

func targetName(target string) string {
	if target == targetLocal {
		return "[local]"
	}
	return target
}
