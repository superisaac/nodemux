package nodemuxcore

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "block_tip",
		Help:      "block tips of each chain/network",
	}, []string{"chain", "network"})

	metricsEndpointBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "endpoint_block_tip",
		Help:      "block tips of each chain/network/endpoint",
	}, []string{"chain", "network", "endpoint"})

	metricsEndpointHealthy = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodemux",
		Name:      "endpoint_healthy",
		Help:      "healthiness of endpoint",
	}, []string{"chain", "network", "endpoint"})
)

func init() {
	prometheus.MustRegister(metricsBlockTip)
	prometheus.MustRegister(metricsEndpointBlockTip)
	prometheus.MustRegister(metricsEndpointHealthy)
}
