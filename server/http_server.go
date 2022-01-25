package server

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/multiplex"
	"net/http"
)

func startServer(bind string, handler http.Handler, tlsConfigs ...*TLSConfig) error {
	var tlsConfig *TLSConfig
	for _, cfg := range tlsConfigs {
		if cfg != nil {
			tlsConfig = cfg
			break
		}
	}

	if tlsConfig != nil {
		return http.ListenAndServeTLS(
			bind,
			tlsConfig.Certfile,
			tlsConfig.Keyfile,
			handler)
	} else {
		return http.ListenAndServe(bind, handler)
	}
}

func startMetricsServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Metrics.Bind
	if bind == "" {
		//bind = "0.0.0.0:9996"
		log.Panicf("metrics bind is empty")
		return
	}

	handler := NewHttpAuthHandler(
		serverCfg.Metrics.Auth,
		promhttp.Handler())
	err := startServer(bind, handler,
		serverCfg.Metrics.TLS,
		serverCfg.TLS)
	if err != nil {
		log.Warnf("start metrics server error %s", err)
	}
}

func startEntrypointServer(rootCtx context.Context, entryCfg *EntrypointConfig, serverCfg *ServerConfig) {
	support, rpcType := multiplex.GetDelegatorFactory().SupportChain(entryCfg.Chain)
	if !support {
		log.Warnf("entry point for chain %s not supported", entryCfg.Chain)
		return
	}
	chain := multiplex.ChainRef{Name: entryCfg.Chain, Network: entryCfg.Network}
	var handler http.Handler
	if rpcType == multiplex.ApiJSONRPC {
		rpc1 := NewRPCRelayer(rootCtx)
		rpc1.chain = chain
		handler = rpc1
	} else if rpcType == multiplex.ApiREST {
		rest1 := NewRESTRelayer(rootCtx)
		rest1.chain = chain
		handler = rest1
	} else {
		// rpcType == multiplex.ApiGraphQL
		graph1 := NewGraphQLRelayer(rootCtx)
		graph1.chain = chain
		handler = graph1
	}
	log.Infof("entrypoint server %s listens at %s", chain, entryCfg.Bind)
	err := startServer(entryCfg.Bind,
		NewHttpAuthHandler(
			serverCfg.Auth, handler),
		entryCfg.TLS, serverCfg.TLS)

	if err != nil {
		log.Println("entry point error ---", err)
	}
}

func StartHTTPServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Bind
	if bind == "" {
		bind = "127.0.0.1:9000"
	}
	log.Infof("start http proxy at %s", bind)
	serverMux := http.NewServeMux()
	serverMux.Handle("/metrics", NewHttpAuthHandler(
		serverCfg.Metrics.Auth,
		promhttp.Handler()))
	serverMux.Handle("/jsonrpc/", NewHttpAuthHandler(
		serverCfg.Auth,
		NewRPCRelayer(rootCtx)))
	serverMux.Handle("/rest/", NewHttpAuthHandler(
		serverCfg.Auth,
		NewRESTRelayer(rootCtx)))
	serverMux.Handle("/graphql/", NewHttpAuthHandler(
		serverCfg.Auth,
		NewGraphQLRelayer(rootCtx)))

	for _, entryCfg := range serverCfg.Entrypoints {
		go startEntrypointServer(rootCtx, entryCfg, serverCfg)
	}

	if serverCfg.Metrics.Bind != "" {
		go startMetricsServer(rootCtx, serverCfg)
	}

	err := startServer(bind, serverMux, serverCfg.TLS)
	if err != nil {
		log.Println("HTTP Server Error - ", err)
		//panic(err)
	}
}

type HttpAuthHandler struct {
	authConfig *AuthConfig
	next       http.Handler
}

func NewHttpAuthHandler(authConfig *AuthConfig, next http.Handler) *HttpAuthHandler {
	return &HttpAuthHandler{authConfig: authConfig, next: next}
}

func (self HttpAuthHandler) TryAuth(r *http.Request) bool {
	if self.authConfig == nil {
		return true
	}

	if self.authConfig.Basic != nil {
		basicAuth := self.authConfig.Basic
		if username, password, ok := r.BasicAuth(); ok {
			if basicAuth.Username == username && basicAuth.Password == password {
				return true
			}
		}
	}

	if self.authConfig.Bearer != nil && self.authConfig.Bearer.Token != "" {
		bearerAuth := self.authConfig.Bearer
		authHeader := r.Header.Get("Authorization")
		expect := fmt.Sprintf("Bearer %s", bearerAuth.Token)
		if authHeader == expect {
			return true
		}
	}

	return false
}

func (self *HttpAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !self.TryAuth(r) {
		w.WriteHeader(401)
		w.Write([]byte("auth failed!\n"))
		return
	}
	self.next.ServeHTTP(w, r)
}
