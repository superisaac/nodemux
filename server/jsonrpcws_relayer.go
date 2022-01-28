package server

import (
	"context"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/jsonrpc/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type JSONRPCWSRelayer struct {
	rootCtx   context.Context
	regex     *regexp.Regexp
	chain     nodemuxcore.ChainRef
	rpcServer *jsonrpchttp.WSServer
}

func NewJSONRPCWSRelayer(rootCtx context.Context) *JSONRPCWSRelayer {
	relayer := &JSONRPCWSRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc\-ws/([^/]+)/([^/]+)/?$`),
	}

	rpcServer := jsonrpchttp.NewWSServer(nil)
	rpcServer.Router.OnMissing(func(req *jsonrpchttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *JSONRPCWSRelayer) delegateRPC(req *jsonrpchttp.RPCRequest) (interface{}, error) {
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
		chain = nodemuxcore.ChainRef{
			Name:    chainName,
			Network: network,
		}
	}

	if !msg.IsRequest() {
		return nil, jsonrpchttp.SimpleHttpResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsonrpc.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		return nil, jsonrpchttp.SimpleHttpResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, m, chain, reqmsg)
	return resmsg, err
}

func (self *JSONRPCWSRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcServer.ServeHTTP(w, r)
} // JSONRPCWSRelayer.ServeHTTP
