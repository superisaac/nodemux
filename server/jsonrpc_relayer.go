package server

import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

// JSONRPC Handler
type JSONRPCRelayer struct {
	rootCtx    context.Context
	acc        *Acc
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
	acc := self.acc

	if acc == nil {
		acc = AccFromContext(r.Context())
		if acc == nil {
			return nil, jlibhttp.SimpleResponse{
				Code: 404,
				Body: []byte("acc not found"),
			}
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

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(acc.Chain.Brand)
	if delegator == nil {
		return nil, jlibhttp.SimpleResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, m, acc.Chain, reqmsg, r)
	return resmsg, err
}

func (self *JSONRPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcHandler.ServeHTTP(w, r)
} // JSONRPCRelayer.ServeHTTP
