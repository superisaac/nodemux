package server

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

var (
	rpcRegex = regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)/?$`)
)

// JSONRPC Handler
type JSONRPCRelayer struct {
	rootCtx   context.Context
	chain     nodemuxcore.ChainRef
	rpcServer *jsonzhttp.Server
}

func NewJSONRPCRelayer(rootCtx context.Context) *JSONRPCRelayer {
	relayer := &JSONRPCRelayer{
		rootCtx: rootCtx,
	}

	rpcServer := jsonzhttp.NewServer()
	rpcServer.Handler.OnMissing(func(req *jsonz.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *JSONRPCRelayer) delegateRPC(req *jsonz.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	if chain.Empty() {
		matches := rpcRegex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			return nil, jsonzhttp.SimpleResponse{
				Code: 404,
				Body: []byte("not found"),
			}
		}
		chainName := matches[1]
		network := matches[2]
		chain = nodemuxcore.ChainRef{Name: chainName, Network: network}
	}

	if !msg.IsRequest() {
		return nil, jsonzhttp.SimpleResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsonz.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		return nil, jsonzhttp.SimpleResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, m, chain, reqmsg)
	return resmsg, err
}

func (self *JSONRPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcServer.ServeHTTP(w, r)
} // JSONRPCRelayer.ServeHTTP
