package main

import "sync/atomic"

type Metrics struct {
	carbonIntensity atomic.Int64
}

func NewMetrics() *Metrics {
	metrics := &Metrics{}
	metrics.carbonIntensity.Store(142)
	return metrics
}

func (m *Metrics) CarbonIntensity() int         { return int(m.carbonIntensity.Load()) }
func (m *Metrics) SetCarbonIntensity(value int) { m.carbonIntensity.Store(int64(value)) }
