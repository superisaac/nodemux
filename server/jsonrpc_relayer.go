package server

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/multiplex"
	"io"
	"net/http"
	"regexp"
)

// JSONRPC Handler
type RPCRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	chain   multiplex.ChainRef
}

func NewRPCRelayer(rootCtx context.Context) *RPCRelayer {
	return &RPCRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/jsonrpc/([^/]+)/([^/]+)/?$`),
	}
}

func (self *RPCRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// only support POST
	if r.Method != "POST" {
		jsonrpc.ErrorResponse(w, r, errors.New("method not allowed"), 405, "Method not allowed")
		return
	}

	chain := self.chain
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			log.Warnf("http url pattern failed")
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		chainName := matches[1]
		network := matches[2]
		chain = multiplex.ChainRef{Name: chainName, Network: network}
	}

	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r.Body)
	if err != nil {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	msg, err := jsonrpc.ParseBytes(buffer.Bytes())
	if err != nil {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	if !msg.IsRequest() {
		jsonrpc.ErrorResponse(w, r, err, 400, "Bad request")
		return
	}

	reqmsg, _ := msg.(*jsonrpc.RequestMessage)
	blcer := multiplex.GetMultiplexer()

	delegator := multiplex.GetDelegatorFactory().GetRPCDelegator(chain.Name)
	if delegator == nil {
		jsonrpc.ErrorResponse(w, r, err, 404, "backend not found")
		return
	}

	resmsg, err := delegator.DelegateRPC(self.rootCtx, blcer, chain, reqmsg)
	if err != nil {
		// put the original http response
		var abnErr *multiplex.AbnormalResponse
		if errors.As(err, &abnErr) {
			origResp := abnErr.Response
			for hn, hvs := range origResp.Header {
				// TODO: filter scan headers
				for _, hv := range hvs {
					w.Header().Add(hn, hv)
				}
			}
			w.WriteHeader(origResp.StatusCode)
			io.Copy(w, origResp.Body)
			return
		}
		jsonrpc.ErrorResponse(w, r, err, 500, "Server error")
		return
	}

	data, err1 := jsonrpc.MessageBytes(resmsg)
	if err1 != nil {
		jsonrpc.ErrorResponse(w, r, err1, 500, "Server error")
		return
	}
	w.Write(data)
} // RPCRelayer.ServeHTTP
