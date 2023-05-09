package server

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/superisaac/nodemux/core"
	"github.com/superisaac/nodemux/ratelimit"
)

var (
	metricsWSPairsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "websocket_pairs_count",
		Help:      "the count of websocket pairs",
	})
)

// Ratelimit collector
var ratelimitDesc = prometheus.NewDesc(
	"nodemux_ratelimit_value",
	"the current ratelimit values",
	[]string{"source"}, nil)

type RatelimitCollector struct {
}

func NewRatelimitCollector() *RatelimitCollector {
	return &RatelimitCollector{}
}

func (self RatelimitCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- ratelimitDesc
}

func (self RatelimitCollector) Collect(ch chan<- prometheus.Metric) {
	m := nodemuxcore.GetMultiplexer()
	if c, ok := m.RedisClient("ratelimit"); ok {
		if values, err := ratelimit.Values(context.Background(), c); err == nil {
			for field, v := range values {
				ch <- prometheus.MustNewConstMetric(
					ratelimitDesc,
					prometheus.GaugeValue,
					float64(v),
					field)
			}
		}
	}
}

func init() {
	prometheus.MustRegister(
		metricsWSPairsCount, NewRatelimitCollector())
}
