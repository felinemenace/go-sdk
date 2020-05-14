// Copyright (c) 2016 - 2020 Sqreen. All Rights Reserved.
// Please refer to our terms for more information:
// https://www.sqreen.io/terms.html

// Package api provides the base data structures of security signals.
// Higher-level signals can be built from there, such as HTTP traces, metrics,
// events, etc.
package api

import "time"

func NewSumMetric(name, source string, started, ended time.Time, interval time.Duration, values map[string]int64) *Metric {
	return NewMetric(name, source, started, newMetricPayload(started, ended, interval, "sum", values))
}

func newMetricPayload(started, ended time.Time, interval time.Duration, kind string, values map[string]int64) *SignalPayload {
	kvArray := make([]MetricValueEntry, 0, len(values))
	for k, v := range values {
		kvArray = append(kvArray, MetricValueEntry{Key: k, Value: v})
	}

	captureIntervalSec := int64(interval / time.Second)

	return NewPayload("metric/2020-01-01T00:00:00.000Z", MetricSignalPayload{
		Type:               "metric",
		CaptureIntervalSec: captureIntervalSec,
		DateStarted:        started,
		DateEnded:          ended,
		Kind:               kind,
		Values:             kvArray,
	})
}

type (
	MetricSignalPayload struct {
		Type               string
		CaptureIntervalSec int64              `json:"capture_interval_s"`
		DateStarted        time.Time          `json:"date_started"`
		DateEnded          time.Time          `json:"date_ended"`
		Kind               string             `json:"kind"`
		Values             []MetricValueEntry `json:"values"`
	}

	MetricValueEntry struct {
		Key   string `json:"key"`
		Value int64  `json:"value"`
	}
)
