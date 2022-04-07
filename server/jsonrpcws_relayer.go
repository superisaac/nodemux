package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"net/url"
	"regexp"
)

var (
	wsPairs = make(map[string]*jlibhttp.WSClient)

	wsRegex = regexp.MustCompile(`^/jsonrpc\-ws/([^/]+)/([^/]+)/?$`)
)

// JSONRPC Handler
type JSONRPCWSRelayer struct {
	rootCtx    context.Context
	chain      nodemuxcore.ChainRef
	rpcHandler *jlibhttp.WSHandler
}

func NewJSONRPCWSRelayer(rootCtx context.Context) *JSONRPCWSRelayer {
	relayer := &JSONRPCWSRelayer{
		rootCtx: rootCtx,
	}

	rpcHandler := jlibhttp.NewWSHandler(rootCtx, nil)
	rpcHandler.Actor.OnClose(func(r *http.Request, s jlibhttp.RPCSession) {
		relayer.onClose(r, s)
	})
	rpcHandler.Actor.OnMissing(func(req *jlibhttp.RPCRequest) (interface{}, error) {
		serverCfg := ServerConfigFromContext(rootCtx)
		ok, err := checkRatelimit(req.HttpRequest(), serverCfg.Ratelimit)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, jlibhttp.SimpleResponse{
				Code: 403,
				Body: []byte("rate limit exceeded!"),
			}
		}
		return relayer.delegateRPC(req)
	})
	relayer.rpcHandler = rpcHandler
	return relayer
}

func (self *JSONRPCWSRelayer) onClose(r *http.Request, s jlibhttp.RPCSession) {
	delete(wsPairs, s.SessionID())
	metricsWSPairsCount.Set(float64(len(wsPairs)))
}

func (self *JSONRPCWSRelayer) delegateRPC(req *jlibhttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	session := req.Session()
	if session == nil {
		return nil, errors.New("request data is not websocket conn")
	}

	if chain.Empty() {
		matches := wsRegex.FindStringSubmatch(r.URL.Path)
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

	m := nodemuxcore.GetMultiplexer()

	if destWs, ok := wsPairs[session.SessionID()]; ok {
		// a existing dest ws conn found, relay the message to it
		err := destWs.Send(self.rootCtx, msg)
		return nil, err
	} else if ep, found := m.SelectWebsocketEndpoint(chain, "", -2); found {
		// the first time a websocket connection connects
		// select an available dest websocket connection
		// make a pair (session, destWs)
		u, err := url.Parse(ep.Config.Url)
		if err != nil {
			return nil, err
		}
		destWs := jlibhttp.NewWSClient(u)
		destWs.OnMessage(func(m jlib.Message) {
			session.Send(m)
		})
		wsPairs[session.SessionID()] = destWs
		metricsWSPairsCount.Set(float64(len(wsPairs)))
		err = destWs.Send(self.rootCtx, msg)
		return nil, err
	} else if msg.IsRequest() {
		// if no dest websocket connection is available and msg is a request message
		// it's still ok to deliver the message to http endpoints
		delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Brand)
		reqmsg, _ := msg.(*jlib.RequestMessage)
		if delegator == nil {
			return nil, jlibhttp.SimpleResponse{
				Code: 404,
				Body: []byte("backend not found"),
			}
		}

		resmsg, err := delegator.DelegateRPC(self.rootCtx, m, chain, reqmsg, r)
		return resmsg, err
	} else {
		// the last way, return back
		return nil, jlibhttp.SimpleResponse{
			Code: 400,
			Body: []byte("no websocket upstreams"),
		}
	}
}

func (self *JSONRPCWSRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcHandler.ServeHTTP(w, r)
} // JSONRPCWSRelayer.ServeHTTP
