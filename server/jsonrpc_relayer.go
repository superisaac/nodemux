package server

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type JSONRPCRelayer struct {
	rootCtx   context.Context
	regex     *regexp.Regexp
	chain     nodemuxcore.ChainRef
	rpcServer *jsonzhttp.Server
}

func NewJSONRPCRelayer(rootCtx context.Context) *JSONRPCRelayer {
	relayer := &JSONRPCRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)/?$`),
	}

	rpcServer := jsonzhttp.NewServer(nil)
	rpcServer.Router.OnMissing(func(req *jsonzhttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *JSONRPCRelayer) delegateRPC(req *jsonzhttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			return nil, jsonzhttp.SimpleHttpResponse{
				Code: 404,
				Body: []byte("not found"),
			}
		}
		chainName := matches[1]
		network := matches[2]
		chain = nodemuxcore.ChainRef{Name: chainName, Network: network}
	}

	if !msg.IsRequest() {
		return nil, jsonzhttp.SimpleHttpResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsonz.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		return nil, jsonzhttp.SimpleHttpResponse{
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
