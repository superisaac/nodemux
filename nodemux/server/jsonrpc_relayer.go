package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"time"
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

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(acc.Chain.Namespace)
	if delegator == nil {
		return nil, jlibhttp.SimpleResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	start := time.Now()
	resmsg, err := delegator.DelegateRPC(self.rootCtx, m, acc.Chain, reqmsg, r)
	// metrics the call time
	delta := time.Now().Sub(start)

	acc.Chain.Log().WithFields(log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": delta.Milliseconds(),
		"account":     acc.Name,
	}).Info("delegate jsonrpc")
	return resmsg, err
}

func (self *JSONRPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcHandler.ServeHTTP(w, r)
} // JSONRPCRelayer.ServeHTTP