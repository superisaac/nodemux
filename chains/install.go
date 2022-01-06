package chains

import (
	"github.com/superisaac/nodeb/balancer"
)

func InstallAdaptors(factory *balancer.DelegatorFactory) {
	// JSON-RPC handlers
	factory.RegisterRPC(NewEthereumChain(),
		"ethereum", "binance-chain", "polygon",
		"okex-token", "huobi-token", "ethereum-classic",
		"cardano-kevm",
	)
	factory.RegisterRPC(NewRippleChain(), "ripple")
	factory.RegisterRPC(NewFilecoinChain(), "filecoin")
	factory.RegisterRPC(NewSolanaChain(), "solana")
	factory.RegisterRPC(NewStarcoinChain(), "starcoin")
	factory.RegisterRPC(NewConfluxChain(), "conflux")
	factory.RegisterRPC(NewPolkadotChain(), "polkadot", "kusama")
	factory.RegisterRPC(NewBitcoinChain(),
		"bitcoin", "litecoin", "dogecoin",
		"bitcoin-cash", "omnicore",
		"dashcoin", "zcash")

	// REST handlers
	factory.RegisterREST(NewTronChain(), "tron-full", "tron-grid")
	factory.RegisterREST(NewEosChain(), "eosio", "enu")
	factory.RegisterREST(NewAlgorandChain(), "algorand")
	factory.RegisterREST(NewKadenaChain(), "kadena")
}
