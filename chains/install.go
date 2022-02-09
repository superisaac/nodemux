package chains

import (
	"github.com/superisaac/nodemux/core"
)

func InstallAdaptors(factory *nodemuxcore.DelegatorFactory) {
	// JSON-RPC handlers
	factory.RegisterRPC(NewEthereumChain(),
		"ethereum", "binance-chain", "polygon",
		"okex-token", "huobi-token", "ethereum-classic",
		"cardano-kevm",
		"fantom-web3",
	)

	factory.RegisterRPC(NewBitcoinChain(),
		"bitcoin", "litecoin", "dogecoin",
		"bitcoin-cash", "omnicore",
		"dashcoin", "zcash")

	factory.RegisterRPC(NewFilecoinChain(), "filecoin")
	factory.RegisterRPC(NewSolanaChain(), "solana")
	factory.RegisterRPC(NewStarcoinChain(), "starcoin")
	factory.RegisterRPC(NewConfluxChain(), "conflux")
	factory.RegisterRPC(NewPolkadotChain(), "polkadot", "kusama")

	// REST handlers
	factory.RegisterREST(NewRippleChain(), "ripple")
	factory.RegisterREST(NewTronChain(), "tron-full", "tron-grid")
	factory.RegisterREST(NewEosChain(), "eosio", "enu")
	factory.RegisterREST(NewLunaChain(), "luna")
	factory.RegisterREST(NewAlgorandChain(), "algorand")
	factory.RegisterREST(NewKadenaChain(), "kadena")

	// GraphQL handlers
	factory.RegisterGraphQL(NewFantomChain(), "fantom")
	factory.RegisterGraphQL(NewCardanoChain(), "cardano")
}

// func init() {
// 	factory := nodemuxcore.GetDelegatorFactory()
// 	InstallAdaptors(factory)
// }
