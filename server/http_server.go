package server

import (
	//"fmt"
	"bytes"
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/balancer"
	"net"
	"net/http"
	"regexp"
)

type HTTPOption struct {
	keyFile  string
	certFile string
}

type HTTPOptionFunc func(opt *HTTPOption)

func WithTLS(certFile string, keyFile string) HTTPOptionFunc {
	return func(opt *HTTPOption) {
		opt.certFile = certFile
		opt.keyFile = keyFile
	}
}

func StartHTTPServer(rootCtx context.Context, bind string, opts ...HTTPOptionFunc) {
	httpOption := &HTTPOption{}
	for _, opt := range opts {
		opt(httpOption)
	}

	log.Infof("start http proxy at %s", bind)
	mux := http.NewServeMux()
	//mux.Handle("/metrics", NewMetricsCollector(rootCtx))
	//mux.Handle("/ws", NewWSServer(rootCtx))
	mux.Handle("/", NewRPCRelayer(rootCtx))

	server := &http.Server{Addr: bind, Handler: mux}
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		panic(err)
	}

	serverCtx, cancelServer := context.WithCancel(rootCtx)
	defer cancelServer()

	go func() {
		for {
			<-serverCtx.Done()
			log.Debugf("http server %s stops", bind)
			listener.Close()
			return
		}
	}()

	if httpOption.certFile != "" && httpOption.keyFile != "" {
		err = server.ServeTLS(listener, httpOption.certFile, httpOption.keyFile)
	} else {
		err = server.Serve(listener)
	}
	if err != nil {
		log.Println("HTTP Server Error - ", err)
		//panic(err)
	}
}

// JSONRPC HTTP Server
type RPCRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
}

func NewRPCRelayer(rootCtx context.Context) *RPCRelayer {
	return &RPCRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^([^/]+)/([^/]+)$`),
	}
}

func (self *RPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only support POST
	if r.Method != "POST" {
		jsonrpc.ErrorResponse(w, r, errors.New("method not allowed"), 405, "Method not allowed")
		return
	}

	matches := self.regex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 3 {
		log.Warnf("http url pattern failed")
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return
	}
	chainName := matches[1]
	network := matches[2]
	chain := balancer.ChainRef{Name: chainName, Network: network}

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
	blcer := balancer.GetBalancer()

	resmsg, err := blcer.RelayMessage(self.rootCtx, chain, reqmsg)
	if err != nil {
		jsonrpc.ErrorResponse(w, r, err, 500, "Server error")
		return
	}

	data, err1 := jsonrpc.MessageBytes(resmsg)
	if err1 != nil {
		jsonrpc.ErrorResponse(w, r, err1, 500, "Server error")
		return
	}
	w.Write(data)
} // ServeHTTP