package nodemuxcore

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "block_tip",
		Help:      "block tips of each chain/network",
	}, []string{"chain"})

	metricsEndpointBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "endpoint_block_tip",
		Help:      "block tips of each chain/network/endpoint",
	}, []string{"chain", "endpoint"})

	metricsEndpointHealthy = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "endpoint_healthy",
		Help:      "healthiness of endpoint",
	}, []string{"chain", "endpoint"})

	metricsEndpointRelayCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "nodemux",
		Name:      "endpoint_relay_count",
		Help:      "the count of endpoint relays",
	}, []string{"endpoint"})
)

func init() {
	prometheus.MustRegister(metricsBlockTip)
	prometheus.MustRegister(metricsEndpointBlockTip)
	prometheus.MustRegister(metricsEndpointHealthy)
	prometheus.MustRegister(metricsEndpointRelayCount)
}
