package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/nmux"
	"io"
	//"net"
	"net/http"
	"regexp"
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
	support, rpcType := nmux.GetDelegatorFactory().SupportChain(entryCfg.Chain)
	if !support {
		log.Warnf("entry point for chain %s not supported", entryCfg.Chain)
		return
	}
	chain := nmux.ChainRef{Name: entryCfg.Chain, Network: entryCfg.Network}
	var handler http.Handler
	if rpcType == "JSONRPC" {
		rpc1 := NewRPCRelayer(rootCtx)
		rpc1.chain = chain
		handler = rpc1
	} else {
		// rpcType == 'REST'
		rest1 := NewRESTRelayer(rootCtx)
		rest1.chain = chain
		handler = rest1
	}
	log.Infof("entrypoint server %s listens at %s", chain, entryCfg.Bind)
	err := startServer(entryCfg.Bind,
		NewHttpAuthHandler(
			serverCfg.Auth, handler),
		entryCfg.TLS, serverCfg.TLS)

	// if serverCfg.TLS != nil {
	// 	err = http.ListenAndServeTLS(
	// 		entryCfg.Bind,
	// 		serverCfg.TLS.Certfile,
	// 		serverCfg.TLS.Keyfile,
	// 		NewHttpAuthHandler(
	// 			serverCfg.Auth, handler))
	// } else {
	// 	err = http.ListenAndServe(entryCfg.Bind,
	// 		NewHttpAuthHandler(
	// 			serverCfg.Auth, handler))
	// }
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
	serverMux.Handle("/jsonrpc", NewHttpAuthHandler(
		serverCfg.Auth,
		NewRPCRelayer(rootCtx)))
	serverMux.Handle("/rest", NewHttpAuthHandler(
		serverCfg.Auth,
		NewRESTRelayer(rootCtx)))

	for _, entryCfg := range serverCfg.Entrypoints {
		go startEntrypointServer(rootCtx, entryCfg, serverCfg)
	}

	if serverCfg.Metrics.Bind != "" {
		go startMetricsServer(rootCtx, serverCfg)
	}

	// server := &http.Server{Addr: bind, Handler: serverMux}
	// listener, err := net.Listen("tcp", bind)
	// if err != nil {
	// 	panic(err)
	// }

	// serverCtx, cancelServer := context.WithCancel(rootCtx)
	// defer cancelServer()
	// go func() {
	// 	for {
	// 		<-serverCtx.Done()
	// 		log.Debugf("http server %s stops", bind)
	// 		listener.Close()
	// 		return
	// 	}
	// }()

	err := startServer(bind, serverMux, serverCfg.TLS)
	// var err error
	// if serverCfg.TLS != nil {
	// 	err = http.ListenAndServeTLS(bind, serverCfg.TLS.Certfile, serverCfg.TLS.Keyfile, serverMux)
	// } else {
	// 	err = http.ListenAndServe(bind, serverMux)
	// }
	if err != nil {
		log.Println("HTTP Server Error - ", err)
		//panic(err)
	}
}

// JSONRPC Handler
type RPCRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	chain   nmux.ChainRef
}

func NewRPCRelayer(rootCtx context.Context) *RPCRelayer {
	return &RPCRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)$`),
	}
}

func (self *RPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only support POST
	if r.Method != "POST" {
		jsonrpc.ErrorResponse(w, r, errors.New("method not allowed"), 405, "Method not allowed")
		return
	}

	chain := self.chain
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			log.Warnf("http url pattern failed")
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		chainName := matches[1]
		network := matches[2]
		chain = nmux.ChainRef{Name: chainName, Network: network}
	}

	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r.Body)
	if err != nil {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	msg, err := jsonrpc.ParseBytes(buffer.Bytes())
	if err != nil {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	if !msg.IsRequest() {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	reqmsg, _ := msg.(*jsonrpc.RequestMessage)
	blcer := nmux.GetMultiplexer()

	delegator := nmux.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		jsonrpc.ErrorResponse(w, r, err, 404, "backend not found")
		return
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, blcer, chain, reqmsg)
	if err != nil {
		// put the original http response
		var abnErr *nmux.AbnormalResponse
		if errors.As(err, &abnErr) {
			origResp := abnErr.Response
			for hn, hvs := range origResp.Header {
				// TODO: filter scan headers
				for _, hv := range hvs {
					w.Header().Add(hn, hv)
				}
			}
			w.WriteHeader(origResp.StatusCode)
			io.Copy(w, origResp.Body)
			return
		}
		jsonrpc.ErrorResponse(w, r, err, 500, "Server error")
		return
	}

	data, err1 := jsonrpc.MessageBytes(resmsg)
	if err1 != nil {
		jsonrpc.ErrorResponse(w, r, err1, 500, "Server error")
		return
	}
	w.Write(data)
} // RPCRelayer.ServeHTTP

// REST Handler
type RESTRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	chain   nmux.ChainRef
}

func NewRESTRelayer(rootCtx context.Context) *RESTRelayer {
	return &RESTRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/rest/([^/]+)/([^/]+)/(.*)$`),
	}
}

func (self *RESTRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := self.chain
	method := r.URL.Path
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 4 {
			log.Warnf("http url pattern failed")
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		chainName := matches[1]
		network := matches[2]
		method = "/" + matches[3]
		chain = nmux.ChainRef{Name: chainName, Network: network}
	}

	blcer := nmux.GetMultiplexer()

	delegator := nmux.GetDelegatorFactory().GetRESTDelegator(chain.Name)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	err := delegator.DelegateREST(self.rootCtx, blcer, chain, method, w, r)
	if err != nil {
		log.Warnf("error delegate rest %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // RESTRelayer.ServeHTTP

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
