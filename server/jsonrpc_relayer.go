package server

import (
	//"bytes"
	"context"
	//"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/jsonrpc/http"
	"github.com/superisaac/nodemux/multiplex"
	//"io"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type RPCRelayer struct {
	rootCtx   context.Context
	regex     *regexp.Regexp
	chain     multiplex.ChainRef
	rpcServer *jsonrpchttp.Server
}

func NewRPCRelayer(rootCtx context.Context) *RPCRelayer {
	relayer := &RPCRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)/?$`),
	}

	rpcServer := jsonrpchttp.NewServer(nil)
	rpcServer.Router.OnMissing(func(req *jsonrpchttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *RPCRelayer) delegateRPC(req *jsonrpchttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			return nil, jsonrpchttp.SimpleHttpResponse{
				Code: 404,
				Body: []byte("not found"),
			}
		}
		chainName := matches[1]
		network := matches[2]
		chain = multiplex.ChainRef{Name: chainName, Network: network}
	}

	if !msg.IsRequest() {
		//jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return nil, jsonrpchttp.SimpleHttpResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsonrpc.RequestMessage)
	blcer := multiplex.GetMultiplexer()

	delegator := multiplex.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		//jsonrpc.ErrorResponse(w, r, err, 404, "backend not found")
		return nil, jsonrpchttp.SimpleHttpResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, blcer, chain, reqmsg)
	return resmsg, err
}

func (self *RPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcServer.ServeHTTP(w, r)
} // RPCRelayer.ServeHTTP
