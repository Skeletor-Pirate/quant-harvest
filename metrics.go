// Copyright 2026 Skeletor-Pirate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
