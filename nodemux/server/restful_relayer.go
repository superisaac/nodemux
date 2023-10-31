package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"

	"github.com/superisaac/nodemux/core"
)

// REST Handler
type RESTRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	acc     *Acc
}

func NewRESTRelayer(rootCtx context.Context) *RESTRelayer {
	return &RESTRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/rest/([^/]+/[^/]+/[^/]+)/(.*)$`),
	}
}

func (self *RESTRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acc := self.acc
	method := r.URL.Path
	if acc == nil {
		acc = AccFromContext(r.Context())
		if acc == nil {
			w.WriteHeader(404)
			w.Write([]byte("acc not found"))
			return
		}

		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 3 {
			requestLog(r).Warnf("http url pattern failed")
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		method = "/" + matches[2]
	}

	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRESTDelegator(acc.Chain.Namespace)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	if r.Method == "POST" || r.Method == "PUT" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			requestLog(r).Warnf("error reading request body %#v", err)
			w.WriteHeader(400)
			w.Write([]byte("bad request"))
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	err := delegator.DelegateREST(self.rootCtx, m, acc.Chain, method, w, r)
	if err != nil {
		requestLog(r).Warnf("error delegate rest %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // RESTRelayer.ServeHTTP
