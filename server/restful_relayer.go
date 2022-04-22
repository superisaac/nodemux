package server

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
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
		regex:   regexp.MustCompile(`^/rest/([^/]+/[^/]+)/(.*)$`),
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

	delegator := nodemuxcore.GetDelegatorFactory().GetRESTDelegator(acc.Chain.Brand)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	err := delegator.DelegateREST(self.rootCtx, m, acc.Chain, method, w, r)
	if err != nil {
		requestLog(r).Warnf("error delegate rest %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // RESTRelayer.ServeHTTP
