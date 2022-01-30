package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsWSPairsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "websocket_pairs_count",
		Help:      "the count of websocket pairs",
	})
)

func init() {
	prometheus.MustRegister(metricsWSPairsCount)	
}
