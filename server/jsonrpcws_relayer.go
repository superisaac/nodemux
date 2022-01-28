package server

import (
	"context"
	"github.com/superisaac/jsoz"
	"github.com/superisaac/jsoz/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type JSONRPCWSRelayer struct {
	rootCtx   context.Context
	regex     *regexp.Regexp
	chain     nodemuxcore.ChainRef
	rpcServer *jsozhttp.WSServer
}

func NewJSONRPCWSRelayer(rootCtx context.Context) *JSONRPCWSRelayer {
	relayer := &JSONRPCWSRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc\-ws/([^/]+)/([^/]+)/?$`),
	}

	rpcServer := jsozhttp.NewWSServer(nil)
	rpcServer.Router.OnMissing(func(req *jsozhttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *JSONRPCWSRelayer) delegateRPC(req *jsozhttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			return nil, jsozhttp.SimpleHttpResponse{
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
		return nil, jsozhttp.SimpleHttpResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsoz.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		return nil, jsozhttp.SimpleHttpResponse{
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
