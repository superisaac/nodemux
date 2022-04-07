package server

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"github.com/superisaac/nodemux/ratelimit"
	"net/http"
)

func requestLog(r *http.Request) *log.Entry {
	return log.WithFields(log.Fields{
		"remoteAddr": r.RemoteAddr,
	})
}

func startServer(rootCtx context.Context, bind string, handler http.Handler, tlsConfigs ...*jlibhttp.TLSConfig) error {
	//ratelimitHandler := NewRatelimitHandler(rootCtx, handler)
	return jlibhttp.ListenAndServe(
		rootCtx, bind,
		//ratelimitHandler,
		handler,
		tlsConfigs...)
}

func startMetricsServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Metrics.Bind
	if bind == "" {
		log.Panicf("metrics bind is empty")
		return
	}

	handler := jlibhttp.NewAuthHandler(
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

func startEntrypointServer(rootCtx context.Context, entryCfg *EntrypointConfig, serverCfg *ServerConfig) {
	support, rpcType := nodemuxcore.GetDelegatorFactory().SupportChain(entryCfg.Chain)
	if !support {
		log.Warnf("entry point for chain %s not supported", entryCfg.Chain)
		return
	}
	chain := nodemuxcore.ChainRef{
		Brand:   entryCfg.Chain,
		Network: entryCfg.Network,
	}
	var handler http.Handler
	if rpcType == nodemuxcore.ApiJSONRPC {
		rpc1 := NewJSONRPCRelayer(rootCtx)
		rpc1.chain = chain
		handler = rpc1
	} else if rpcType == nodemuxcore.ApiJSONRPCWS {
		rpc1 := NewJSONRPCWSRelayer(rootCtx)
		rpc1.chain = chain
		handler = rpc1
	} else if rpcType == nodemuxcore.ApiREST {
		rest1 := NewRESTRelayer(rootCtx)
		rest1.chain = chain
		handler = rest1
	} else {
		graph1 := NewGraphQLRelayer(rootCtx)
		graph1.chain = chain
		handler = graph1
	}
	log.Infof("entrypoint server %s listens at %s", chain, entryCfg.Bind)

	err := startServer(rootCtx, entryCfg.Bind,
		handlerChains(
			rootCtx,
			serverCfg.Auth,
			handler),
		entryCfg.TLS, serverCfg.TLS)

	if err != nil {
		log.Println("entry point error ---", err)
	}
}

func handlerChains(rootCtx context.Context, authCfg *jlibhttp.AuthConfig, next http.Handler) http.Handler {
	h := NewRatelimitHandler(rootCtx, next)
	h1 := jlibhttp.NewAuthHandler(authCfg, h)
	return h1
}

func StartHTTPServer(rootCtx context.Context, serverCfg *ServerConfig) {
	bind := serverCfg.Bind
	if bind == "" {
		bind = "127.0.0.1:9000"
	}
	log.Infof("start http proxy at %s", bind)

	var adminAuth *jlibhttp.AuthConfig
	if serverCfg.Admin != nil {
		adminAuth = serverCfg.Admin.Auth
	}

	rootCtx = serverCfg.AddTo(rootCtx)
	serverMux := http.NewServeMux()
	serverMux.Handle("/metrics", handlerChains(
		rootCtx,
		serverCfg.Metrics.Auth,
		promhttp.Handler()))

	serverMux.Handle("/admin", handlerChains(
		rootCtx,
		adminAuth,
		NewAdminHandler()))

	serverMux.Handle("/jsonrpc/", handlerChains(
		rootCtx,
		serverCfg.Auth,
		NewJSONRPCRelayer(rootCtx)))

	serverMux.Handle("/jsonrpc-ws/", handlerChains(
		rootCtx,
		serverCfg.Auth,
		NewJSONRPCWSRelayer(rootCtx)))

	serverMux.Handle("/rest/", handlerChains(
		rootCtx,
		serverCfg.Auth,
		NewRESTRelayer(rootCtx)))
	serverMux.Handle("/graphql/", handlerChains(
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

// handle ratelimit
type RatelimitHandler struct {
	rootCtx context.Context
	next    http.Handler
}

func NewRatelimitHandler(rootCtx context.Context, next http.Handler) *RatelimitHandler {
	return &RatelimitHandler{
		rootCtx: rootCtx,
		next:    next,
	}
}

func (self *RatelimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverCfg := ServerConfigFromContext(self.rootCtx)
	ok, err := checkRatelimit(r, serverCfg.Ratelimit)
	if err != nil {
		requestLog(r).Errorf("error while checking ratelimit %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	} else if !ok {
		w.WriteHeader(403)
		w.Write([]byte("rate limit exceeded!"))
	} else {
		self.next.ServeHTTP(w, r)
	}
}

func checkRatelimit(r *http.Request, ratelimitCfg RatelimitConfig) (bool, error) {
	m := nodemuxcore.GetMultiplexer()
	if c, ok := m.RedisClient("ratelimit"); ok {
		// per user based ratelimit
		if v := r.Context().Value("authInfo"); r != nil {
			if authInfo, ok := v.(*jlibhttp.AuthInfo); ok && authInfo != nil && authInfo.Settings != nil {
				limit := ratelimitCfg.User

				// check against usersettings for ratelimit
				var settingsT struct {
					Ratelimit int
				}
				err := jlib.DecodeInterface(authInfo.Settings, &settingsT)
				if err != nil {
					return false, err
				} else if settingsT.Ratelimit > 0 {
					limit = settingsT.Ratelimit
				} else {
					log.Warnf("user settings.Ratelimit %d <= 0", settingsT.Ratelimit)
				}

				return ratelimit.Incr(
					r.Context(),
					c,
					"u"+authInfo.Username,
					limit)

			}
		}
		// per IP based ratelimit
		return ratelimit.Incr(
			r.Context(),
			c, r.RemoteAddr, ratelimitCfg.IP)
	}
	return true, nil
}
