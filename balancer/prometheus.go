package balancer

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodeb",
		Name:      "block_tip",
		Help:      "block tips of each chain/network",
	}, []string{"chain", "network"})

	metricsEndpointBlockTip = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "nodeb",
		Name:      "endpoint_block_tip",
		Help:      "block tips of each chain/network/endpoint",
	}, []string{"chain", "network", "endpoint"})
)

func init() {
	prometheus.MustRegister(metricsBlockTip)

	prometheus.MustRegister(metricsEndpointBlockTip)

}
