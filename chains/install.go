package chains

import (
	"github.com/superisaac/nodeb/balancer"
)

func InstallAdaptors(balancer *balancer.Balancer) {
	balancer.Register(NewEthereumChain(),
		"ethereum", "binance-chain", "polygon", "okex-token", "huobi-token", "ethereum-classic")
}
