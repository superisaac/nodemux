package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/jsoff/net"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"net/url"
)

var (
	wsPairs = make(map[string]*jsoffnet.WSClient)
)

// JSONRPC Handler
type JSONRPCWSRelayer struct {
	rootCtx    context.Context
	acc        *Acc
	rpcHandler *jsoffnet.WSHandler
}

func NewJSONRPCWSRelayer(rootCtx context.Context) *JSONRPCWSRelayer {
	relayer := &JSONRPCWSRelayer{
		rootCtx: rootCtx,
	}

	rpcHandler := jsoffnet.NewWSHandler(rootCtx, nil)
	rpcHandler.Actor.OnClose(func(s jsoffnet.RPCSession) {
		relayer.onClose(s)
	})
	rpcHandler.Actor.OnMissing(func(req *jsoffnet.RPCRequest) (interface{}, error) {
		r := req.HttpRequest()
		acc := AccFromContext(r.Context())
		accName := ""
		var ratelimit RatelimitConfig
		if acc != nil {
			accName = acc.Config.Username
			if accName == "" {
				accName = acc.Name
			}
			ratelimit = acc.Config.Ratelimit
		} else {
			serverCfg := ServerConfigFromContext(rootCtx)
			ratelimit = serverCfg.Ratelimit
		}
		ok, err := checkRatelimit(r, accName, ratelimit, true)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, jsoffnet.SimpleResponse{
				Code: 429,
				Body: []byte("rate limit exceeded!"),
			}
		}
		return relayer.delegateRPC(req)
	})
	relayer.rpcHandler = rpcHandler
	return relayer
}

func (h *JSONRPCWSRelayer) onClose(s jsoffnet.RPCSession) {
	delete(wsPairs, s.SessionID())
	metricsWSPairsCount.Set(float64(len(wsPairs)))
}

func (h *JSONRPCWSRelayer) delegateRPC(req *jsoffnet.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()

	session := req.Session()
	if session == nil {
		return nil, errors.New("request data is not websocket conn")
	}

	acc := h.acc
	if acc == nil {
		acc = AccFromContext(r.Context())
		if acc == nil {
			return nil, jsoffnet.SimpleResponse{
				Code: 404,
				Body: []byte("acc not found"),
			}
		}
	}

	m := nodemuxcore.GetMultiplexer()

	if destWs, ok := wsPairs[session.SessionID()]; ok {
		// a existing dest ws conn found, relay the message to it
		err := destWs.Send(h.rootCtx, msg)
		return nil, err
	} else if ep, found := m.SelectWebsocketEndpoint(acc.Chain, "", -2); found {
		// the first time a websocket connection connects
		// select an available dest websocket connection
		// make a pair (session, destWs)
		u, err := url.Parse(ep.Config.StreamingUrl)
		if err != nil {
			return nil, err
		}
		destWs := jsoffnet.NewWSClient(u)
		destWs.OnMessage(func(m jsoff.Message) {
			session.Send(m)
		})
		destWs.OnClose(func() {
			h.onClose(session)
		})
		wsPairs[session.SessionID()] = destWs
		metricsWSPairsCount.Set(float64(len(wsPairs)))
		err = destWs.Send(h.rootCtx, msg)
		return nil, err
	} else if msg.IsRequest() {
		// if no dest websocket connection is available and msg is a request message
		// it's still ok to deliver the message to http endpoints
		delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(acc.Chain.Namespace)
		reqmsg, _ := msg.(*jsoff.RequestMessage)
		if delegator == nil {
			return nil, jsoffnet.SimpleResponse{
				Code: 404,
				Body: []byte("backend not found"),
			}
		}

		resmsg, err := delegator.DelegateRPC(h.rootCtx, m, acc.Chain, reqmsg, r)
		return resmsg, err
	} else {
		// the last way, return back
		return nil, jsoffnet.SimpleResponse{
			Code: 400,
			Body: []byte("no websocket upstreams"),
		}
	}
}

func (h *JSONRPCWSRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rpcHandler.ServeHTTP(w, r)
} // JSONRPCWSRelayer.ServeHTTP
