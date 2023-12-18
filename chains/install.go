package chains

// list of blockchains: https://www.technoduet.com/a-comprehensive-list-of-blockchain-platforms/

import (
	"github.com/superisaac/nodemux/core"
)

func InstallAdaptors(factory *nodemuxcore.DelegatorFactory) {
	// JSON-RPC handlers
	factory.RegisterRPC(NewWeb3Chain(),
		"ethereum", "binance-chain", "polygon",
		"okex-token", "huobi-token", "ethereum-classic",
		"cardano-kevm",
		"fantom-web3",
	)

	factory.RegisterRPC(NewBitcoinChain(),
		"bitcoin", "litecoin", "dogecoin",
		"bitcoin-cash", "omnicore",
		"dashcoin", "zcash")

	factory.RegisterRPC(NewMoneroChain(), "monero")
	factory.RegisterRPC(NewFilecoinChain(), "filecoin")
	factory.RegisterRPC(NewHandshakeChain(), "handshake")
	factory.RegisterRPC(NewCasperChain(), "casper")
	factory.RegisterRPC(NewSolanaChain(), "solana")
	factory.RegisterRPC(NewStarcoinChain(), "starcoin")
	factory.RegisterRPC(NewConfluxChain(), "conflux")
	factory.RegisterRPC(NewPolkadotChain(), "polkadot", "kusama")
	factory.RegisterRPC(NewNearChain(), "near")
	factory.RegisterRPC(NewSuiChain(), "sui")
	factory.RegisterRPC(NewEOSRPC(), "eosio-rpc", "enu")

	// REST handlers
	factory.RegisterREST(NewRippleChain(), "ripple")
	factory.RegisterREST(NewTronChain(), "tron-full", "tron-grid")
	factory.RegisterREST(NewEOSAPI(), "eosio", "enu")
	factory.RegisterREST(NewCosmosChain(), "cosmos", "luna")
	factory.RegisterREST(NewAlgorandChain(), "algorand")
	factory.RegisterREST(NewKadenaChain(), "kadena")
	factory.RegisterREST(NewSubstrateAPI(), "substrate")
	factory.RegisterREST(NewAptosAPI(), "aptos")

	// GraphQL handlers
	factory.RegisterGraphQL(NewFantomChain(), "fantom")
	factory.RegisterGraphQL(NewCardanoChain(), "cardano")
	factory.RegisterGraphQL(NewMinaChain(), "mina")
}

// func init() {
// 	factory := nodemuxcore.GetDelegatorFactory()
// 	InstallAdaptors(factory)
// }
