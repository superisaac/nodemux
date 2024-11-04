package server

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff/net"
	"net/http"
)

func requestLog(r *http.Request) *log.Entry {
	return log.WithFields(log.Fields{
		"remoteAddr": r.RemoteAddr,
	})
}

func startServer(rootCtx context.Context, bind string, handler http.Handler, tlsConfigs ...*jsoffnet.TLSConfig) error {
	return jsoffnet.ListenAndServe(
		rootCtx, bind,
		handler,
		tlsConfigs...)
}

func startMetricsServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Metrics.Bind
	if bind == "" {
		log.Panicf("metrics bind is empty")
		return
	}

	handler := jsoffnet.NewAuthHandler(
		serverCfg.Metrics.Auth,
		promhttp.Handler())
	err := startServer(
		rootCtx, bind, handler,
		serverCfg.Metrics.TLS,
		serverCfg.TLS)
	if err != nil {
		log.Warnf("start metrics server error %s", err)
	}
}

func adminHandler(_ context.Context, authCfg *jsoffnet.AuthConfig, next http.Handler) http.Handler {
	h1 := jsoffnet.NewAuthHandler(authCfg, next)
	return h1
}

func relayHandler(rootCtx context.Context, authCfg *jsoffnet.AuthConfig, next http.Handler) http.Handler {
	h0 := NewRatelimitHandler(rootCtx, next)
	h1 := NewAccHandler(rootCtx, h0)
	h2 := jsoffnet.NewAuthHandler(authCfg, h1)
	return h2
}

func StartHTTPServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Bind
	if bind == "" {
		bind = "127.0.0.1:9000"
	}
	log.Infof("start http proxy at %s", bind)

	var adminAuth *jsoffnet.AuthConfig
	if serverCfg.Admin != nil {
		adminAuth = serverCfg.Admin.Auth
	}

	rootCtx = serverCfg.AddTo(rootCtx)
	serverMux := http.NewServeMux()
	serverMux.Handle("/metrics", adminHandler(
		rootCtx,
		serverCfg.Metrics.Auth,
		promhttp.Handler()))

	serverMux.Handle("/admin", adminHandler(
		rootCtx,
		adminAuth,
		NewAdminHandler()))

	serverMux.Handle("/jsonrpc/", relayHandler(
		rootCtx,
		serverCfg.Auth,
		NewJSONRPCRelayer(rootCtx)))

	serverMux.Handle("/jsonrpc-ws/", relayHandler(
		rootCtx,
		serverCfg.Auth,
		NewJSONRPCWSRelayer(rootCtx)))

	serverMux.Handle("/rest/", relayHandler(
		rootCtx,
		serverCfg.Auth,
		NewRESTRelayer(rootCtx)))
	serverMux.Handle("/graphql/", relayHandler(
		rootCtx,
		serverCfg.Auth,
		NewGraphQLRelayer(rootCtx)))

	for _, entryCfg := range serverCfg.Entrypoints {
		go startEntrypointServer(rootCtx, entryCfg, serverCfg)
	}

	if serverCfg.Metrics.Bind != "" {
		go startMetricsServer(rootCtx, serverCfg)
	}

	err := startServer(rootCtx, bind, serverMux, serverCfg.TLS)
	if err != nil {
		log.Println("HTTP Server Error - ", err)
		//panic(err)
	}
}
