// Copyright © 2018 Joel Baranick <jbaranick@gmail.com>
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

package client

import (
	"github.com/hashicorp/go-cleanhttp"
	"github.com/kadaan/consulate/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	inFlightGauge prometheus.Gauge
	counter       *prometheus.CounterVec
	histVec       *prometheus.HistogramVec
)

func init() {
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "client_in_flight_requests",
		Help: "A gauge of in-flight requests for the wrapped client.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "client_api_requests_total",
			Help: "A counter for requests from the wrapped client.",
		},
		[]string{"code", "method"},
	)

	histVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_milliseconds",
			Help:    "A histogram of request latencies.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"code", "method"},
	)

	prometheus.MustRegister(counter, histVec, inFlightGauge)
}

// CreateClient creates a new http.Client
func CreateClient(c *config.ClientConfig) *http.Client {
	transport := cleanhttp.DefaultPooledTransport()
	transport.MaxIdleConns = c.QueryMaxIdleConnectionCount
	transport.IdleConnTimeout = c.QueryIdleConnectionTimeout

	roundTripper := promhttp.InstrumentRoundTripperInFlight(inFlightGauge,
		promhttp.InstrumentRoundTripperCounter(counter,
			promhttp.InstrumentRoundTripperDuration(histVec, transport),
		),
	)

	client := &http.Client{
		Transport: roundTripper,
		Timeout:   c.QueryTimeout,
	}

	return client
}
