package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

func startEntrypointServer(rootCtx context.Context, entryCfg EntrypointConfig, serverCfg *ServerConfig) {
	acccfg, ok := serverCfg.Accounts[entryCfg.Account]
	if !ok {
		log.Panicf("account %s not found", entryCfg.Account)
		return
	}

	acc := NewAccFromConfig(entryCfg.Account, acccfg)

	support, rpcType := nodemuxcore.GetDelegatorFactory().SupportChain(acc.Chain.Brand)
	if !support {
		log.Warnf("entry point for chain %s not supported", acc.Chain)
		return
	}
	var handler http.Handler
	if rpcType == nodemuxcore.ApiJSONRPC {
		rpc1 := NewJSONRPCRelayer(rootCtx)
		rpc1.acc = acc
		handler = rpc1
	} else if rpcType == nodemuxcore.ApiJSONRPCWS {
		rpc1 := NewJSONRPCWSRelayer(rootCtx)
		rpc1.acc = acc
		handler = rpc1
	} else if rpcType == nodemuxcore.ApiREST {
		rest1 := NewRESTRelayer(rootCtx)
		rest1.acc = acc
		handler = rest1
	} else {
		graph1 := NewGraphQLRelayer(rootCtx)
		graph1.acc = acc
		handler = graph1
	}
	log.Infof("entrypoint server %s listens at %s", acc.Chain, entryCfg.Bind)

	err := startServer(rootCtx, entryCfg.Bind,
		relayHandler(
			rootCtx,
			serverCfg.Auth,
			handler),
		entryCfg.TLS, serverCfg.TLS)
	if err != nil {
		log.Println("entry point error ---", err)
	}
}
