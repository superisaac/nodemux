package server

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type JSONRPCWSRelayer struct {
	rootCtx   context.Context
	regex     *regexp.Regexp
	chain     nodemuxcore.ChainRef
	rpcServer *jsonzhttp.WSServer

	pairs map[*http.Request]*jsonzhttp.WSClient
}

func NewJSONRPCWSRelayer(rootCtx context.Context) *JSONRPCWSRelayer {
	relayer := &JSONRPCWSRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc\-ws/([^/]+)/([^/]+)/?$`),
		pairs:   make(map[*http.Request]*jsonzhttp.WSClient),
	}

	rpcServer := jsonzhttp.NewWSServer(nil)
	rpcServer.Router.OnClose(func(r *http.Request) {
		relayer.onClose(r)
	})
	rpcServer.Router.OnMissing(func(req *jsonzhttp.RPCRequest) (interface{}, error) {
		return relayer.delegateRPC(req)
	})
	relayer.rpcServer = rpcServer
	return relayer
}

func (self *JSONRPCWSRelayer) onClose(r *http.Request) {
	// FIXME: since map doesn't return if or not the deletion is
	// success, just decrementing a value may not be accurate.
	delete(self.pairs, r)
	metricsWSPairsCount.Set(float64(len(self.pairs)))
}

func (self *JSONRPCWSRelayer) delegateRPC(req *jsonzhttp.RPCRequest) (interface{}, error) {
	r := req.HttpRequest()
	msg := req.Msg()
	chain := self.chain
	data := req.Data()
	if data == nil {
		return nil, errors.New("request data is nil")
	}
	srcWs, ok := data.(*websocket.Conn)
	if !ok {
		return nil, errors.New("request data is not websocket conn")
	}

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
		chain = nodemuxcore.ChainRef{
			Name:    chainName,
			Network: network,
		}
	}

	m := nodemuxcore.GetMultiplexer()

	if destWs, ok := self.pairs[r]; ok {
		// a existing dest ws conn found, relay the message to it
		err := destWs.Send(self.rootCtx, msg)
		return nil, err
	} else if ep, found := m.SelectWebsocketEndpoint(chain, "", -2); found {
		// the first time a websocket connection connects
		// select an available dest websocket connection
		// make a pair (srcWs, destWs)
		destWs := jsonzhttp.NewWSClient(ep.Config.Url)
		destWs.OnMessage(func(m jsonz.Message) {
			err := self.rpcServer.SendMessage(srcWs, m)
			if err != nil {
				m.Log().Warnf("send message error %s", err)
			}
		})
		self.pairs[r] = destWs
		metricsWSPairsCount.Set(float64(len(self.pairs)))
		err := destWs.Send(self.rootCtx, msg)
		return nil, err
	} else if msg.IsRequest() {
		// if no dest websocket connection is available and msg is a request message
		// it's still ok to deliver the message to http endpoints
		delegator := nodemuxcore.GetDelegatorFactory().GetRPCDelegator(chain.Name)
		reqmsg, _ := msg.(*jsonz.RequestMessage)
		if delegator == nil {
			return nil, jsonzhttp.SimpleHttpResponse{
				Code: 404,
				Body: []byte("backend not found"),
			}
		}

		resmsg, err := delegator.DelegateRPC(self.rootCtx, m, chain, reqmsg)
		return resmsg, err
	} else {
		// the last way, return back
		return nil, jsonzhttp.SimpleHttpResponse{
			Code: 400,
			Body: []byte("no websocket upstreams"),
		}
	}
}

func (self *JSONRPCWSRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.rpcServer.ServeHTTP(w, r)
} // JSONRPCWSRelayer.ServeHTTP
