// Copyright 2021 coord_e
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  	 http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"context"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/coord-e/linux_dvb_exporter/linux_dvb/frontend"
)

const namespace = "dvb"

type Exporter struct {
	ctx    context.Context
	logger log.Logger

	status                  *prometheus.Desc
	signal_strength_decibel *prometheus.Desc
	signal_strength_ratio   *prometheus.Desc
	cnr_decibel             *prometheus.Desc
	cnr_ratio               *prometheus.Desc
	pre_error_bytes_count   *prometheus.Desc
	pre_total_bytes_count   *prometheus.Desc
	post_error_bytes_count  *prometheus.Desc
	post_total_bytes_count  *prometheus.Desc
	error_block_count       *prometheus.Desc
	total_block_count       *prometheus.Desc
}

// Verify if Exporter implements prometheus.Collector
var _ prometheus.Collector = (*Exporter)(nil)

func New(ctx context.Context, logger log.Logger) *Exporter {
	const subsystem = "frontend"

	return &Exporter{
		ctx:    ctx,
		logger: logger,
		status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Status information about the DVB frontend devices.",
			[]string{"status", "adapter", "frontend"}, nil),
		signal_strength_decibel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "signal_strength_decibels"),
			"Signal strength level at the analog part of the tuner or of the demod.",
			[]string{"adapter", "frontend"}, nil),
		signal_strength_ratio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "signal_strength_ratio"),
			"Signal strength level at the analog part of the tuner or of the demod.",
			[]string{"adapter", "frontend"}, nil),
		cnr_decibel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "CNR_decibels"),
			"Signal to Noise ratio for the main carrier.",
			[]string{"adapter", "frontend"}, nil),
		cnr_ratio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "CNR_ratio"),
			"Signal to Noise ratio for the main carrier.",
			[]string{"adapter", "frontend"}, nil),
		pre_error_bytes_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pre_error_bytes_total"),
			"Total number of error bytes before the inner code.",
			[]string{"adapter", "frontend"}, nil),
		pre_total_bytes_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pre_bytes_total"),
			"Total number of bytes received before the inner code.",
			[]string{"adapter", "frontend"}, nil),
		post_error_bytes_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "post_error_bytes_total"),
			"Total number of error bytes after the inner code.",
			[]string{"adapter", "frontend"}, nil),
		post_total_bytes_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "post_bytes_total"),
			"Total number of bytes received after the inner code.",
			[]string{"adapter", "frontend"}, nil),
		error_block_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "error_blocks_total"),
			"Total number of error blocks.",
			[]string{"adapter", "frontend"}, nil),
		total_block_count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "blocks_total"),
			"Total number of received blocks.",
			[]string{"adapter", "frontend"}, nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.status
	ch <- e.signal_strength_decibel
	ch <- e.signal_strength_ratio
	ch <- e.cnr_decibel
	ch <- e.cnr_ratio
	ch <- e.pre_error_bytes_count
	ch <- e.pre_total_bytes_count
	ch <- e.post_error_bytes_count
	ch <- e.post_total_bytes_count
	ch <- e.error_block_count
	ch <- e.total_block_count
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	adapters, err := listAllAdapters()
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to list adapters", "err", err)
		return
	}

	for _, adapter := range adapters {
		fes, err := listAllFrontends(adapter)
		if err != nil {
			level.Error(e.logger).Log("msg", "failed to list frontends", "err", err, "adapter", adapter)
			continue
		}

		for _, idx := range fes {
			if err := e.collectFromFrontend(ch, adapter, idx); err != nil {
				level.Error(e.logger).Log("msg", "failed to collect from frontend", "err", err, "adapter", adapter, "frontend", idx)
				continue
			}
		}
	}
}

func (e *Exporter) collectFromFrontend(ch chan<- prometheus.Metric, adapter uint, idx uint) error {
	fe, err := frontend.Open(adapter, idx)
	if err != nil {
		return err
	}

	defer fe.Close()

	status, err := fe.ReadStatus()
	if err != nil {
		return errors.Wrap(err, "failed to read status")
	}

	adapterStr := strconv.Itoa(int(adapter))
	frontendStr := strconv.Itoa(int(idx))

	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.HasSignal), "has_signal", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.HasCarrier), "has_carrier", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.HasViterbi), "has_viterbi", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.HasSync), "has_sync", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.HasLock), "has_lock", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.Timedout), "timedout", adapterStr, frontendStr)
	ch <- prometheus.MustNewConstMetric(e.status, prometheus.UntypedValue, boolToValue(status.Reinit), "reinit", adapterStr, frontendStr)

	stats, err := fe.GetStats()
	if err != nil {
		return errors.Wrap(err, "failed to get stats")
	}

	if stats.SignalStrength.Decibel != nil {
		ch <- prometheus.MustNewConstMetric(e.signal_strength_decibel, prometheus.GaugeValue, *stats.SignalStrength.Decibel, adapterStr, frontendStr)
	}

	if stats.SignalStrength.Ratio != nil {
		ch <- prometheus.MustNewConstMetric(e.signal_strength_ratio, prometheus.GaugeValue, *stats.SignalStrength.Ratio, adapterStr, frontendStr)
	}

	if stats.CNR.Decibel != nil {
		ch <- prometheus.MustNewConstMetric(e.cnr_decibel, prometheus.GaugeValue, *stats.CNR.Decibel, adapterStr, frontendStr)
	}

	if stats.CNR.Ratio != nil {
		ch <- prometheus.MustNewConstMetric(e.cnr_ratio, prometheus.GaugeValue, *stats.CNR.Ratio, adapterStr, frontendStr)
	}

	if stats.PreErrorBitCount != nil {
		ch <- prometheus.MustNewConstMetric(e.pre_error_bytes_count, prometheus.CounterValue, float64(*stats.PreErrorBitCount)/8, adapterStr, frontendStr)
	}

	if stats.PreTotalBitCount != nil {
		ch <- prometheus.MustNewConstMetric(e.pre_total_bytes_count, prometheus.CounterValue, float64(*stats.PreTotalBitCount)/8, adapterStr, frontendStr)
	}

	if stats.PostErrorBitCount != nil {
		ch <- prometheus.MustNewConstMetric(e.post_error_bytes_count, prometheus.CounterValue, float64(*stats.PostErrorBitCount)/8, adapterStr, frontendStr)
	}

	if stats.PostTotalBitCount != nil {
		ch <- prometheus.MustNewConstMetric(e.post_total_bytes_count, prometheus.CounterValue, float64(*stats.PostTotalBitCount)/8, adapterStr, frontendStr)
	}

	if stats.ErrorBlockCount != nil {
		ch <- prometheus.MustNewConstMetric(e.error_block_count, prometheus.CounterValue, float64(*stats.ErrorBlockCount), adapterStr, frontendStr)
	}

	if stats.TotalBlockCount != nil {
		ch <- prometheus.MustNewConstMetric(e.total_block_count, prometheus.CounterValue, float64(*stats.TotalBlockCount), adapterStr, frontendStr)
	}

	return nil
}

func boolToValue(b bool) float64 {
	if b {
		return 1.0
	} else {
		return 0.0
	}
}
