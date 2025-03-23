package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/jsoff/net"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"time"
)

// JSONRPC Handler
type JSONRPCRelayer struct {
	rootCtx    context.Context
	acc        *Acc
	rpcHandler *jsoffnet.Http1Handler
}

func NewJSONRPCRelayer(rootCtx context.Context) *JSONRPCRelayer {
	relayer := &JSONRPCRelayer{
		rootCtx: rootCtx,
	}

	rpcHandler := jsoffnet.NewHttp1Handler(nil)
	rpcHandler.Actor.OnMissing(func(req *jsoffnet.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcHandler = rpcHandler
	return relayer
}

func (h *JSONRPCRelayer) delegateRPC(req *jsoffnet.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	acc := h.acc

	if acc == nil {
		acc = AccFromContext(r.Context())
		if acc == nil {
			return nil, jsoffnet.SimpleResponse{
				Code: 404,
				Body: []byte("account not found"),
			}
		}
	}

	if !msg.IsRequest() {
		return nil, jsoffnet.SimpleResponse{
			Code: 400,
			Body: []byte("bad request"),
		}
	}

	reqmsg, _ := msg.(*jsoff.RequestMessage)
	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(acc.Chain.Namespace)
	if delegator == nil {
		return nil, jsoffnet.SimpleResponse{
			Code: 404,
			Body: []byte("backend not found"),
		}
	}

	start := time.Now()
	if ep := m.SelectEndpointFromHttp(acc.Chain, reqmsg.Method, r); ep != nil {
		resmsg, err := m.CallEndpointRPC(h.rootCtx, ep, reqmsg)
		acc.Chain.Log().WithFields(log.Fields{
			"method":      reqmsg.Method,
			"timeSpentMS": time.Since(start).Milliseconds(),
			"account":     acc.Name,
			"through":     ep.Name,
		}).Info("direct delegate jsonrpc")
		return resmsg, err
	} else {
		resmsg, err := delegator.DelegateRPC(h.rootCtx, m, acc.Chain, reqmsg, r)
		// metrics the call time
		acc.Chain.Log().WithFields(log.Fields{
			"method":      reqmsg.Method,
			"timeSpentMS": time.Since(start).Milliseconds(),
			"account":     acc.Name,
		}).Info("delegate jsonrpc")
		return resmsg, err
	}
}

func (h *JSONRPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rpcHandler.ServeHTTP(w, r)
} // JSONRPCRelayer.ServeHTTP
