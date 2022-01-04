package chains

import (
	"github.com/superisaac/nodeb/balancer"
)

func InstallAdaptors(balancer *balancer.Balancer) {
	balancer.RegisterRPC(NewEthereumChain(),
		"ethereum", "binance-chain", "polygon", "okex-token", "huobi-token", "ethereum-classic")

	balancer.RegisterRPC(NewFilecoinChain(), "filecoin")
	balancer.RegisterRPC(NewSolanaChain(), "solana")
	balancer.RegisterRPC(NewStarcoinChain(), "starcoin")
	balancer.RegisterRPC(NewPolkadotChain(), "polkadot", "kusama")

	balancer.RegisterRPC(NewBitcoinChain(),
		"bitcoin", "litecoin", "dogecoin", "dashcoin", "zcash")
}
