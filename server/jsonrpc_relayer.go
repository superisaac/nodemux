package server

import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

var (
	rpcRegex = regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)/?$`)
)

// JSONRPC Handler
type JSONRPCRelayer struct {
	rootCtx    context.Context
	chain      nodemuxcore.ChainRef
	rpcHandler *jlibhttp.H1Handler
}

func NewJSONRPCRelayer(rootCtx context.Context) *JSONRPCRelayer {
	relayer := &JSONRPCRelayer{
		rootCtx: rootCtx,
	}

	rpcHandler := jlibhttp.NewH1Handler(nil)
	rpcHandler.Actor.OnMissing(func(req *jlibhttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcHandler = rpcHandler
	return relayer
}

func (self *JSONRPCRelayer) delegateRPC(req *jlibhttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	if chain.Empty() {
		matches := rpcRegex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			return nil, jlibhttp.SimpleResponse{
				Code: 404,
				Body: []byte("not found"),
			}
		}
		brand := matches[1]
		network := matches[2]
		chain = nodemuxcore.ChainRef{
			Brand:   brand,
			Network: network,
		}
	}

	if !msg.IsRequest() {
		return nil, jlibhttp.SimpleResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jlib.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Brand)
	if delegator == nil {
		return nil, jlibhttp.SimpleResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, m, chain, reqmsg, r)
	return resmsg, err
}

func (self *JSONRPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcHandler.ServeHTTP(w, r)
} // JSONRPCRelayer.ServeHTTP
