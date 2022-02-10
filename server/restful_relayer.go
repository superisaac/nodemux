package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// REST Handler
type RESTRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	chain   nodemuxcore.ChainRef
}

func NewRESTRelayer(rootCtx context.Context) *RESTRelayer {
	return &RESTRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/rest/([^/]+)/([^/]+)/(.*)$`),
	}
}

func (self *RESTRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := self.chain
	method := r.URL.Path
	if chain.Empty() {
		matches := self.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) < 4 {
			log.Warnf("http url pattern failed")
			w.WriteHeader(404)
			w.Write([]byte("not found"))
			return
		}
		brand := matches[1]
		network := matches[2]
		method = "/" + matches[3]
		chain = nodemuxcore.ChainRef{
			Brand:   brand,
			Network: network}
	}

	m := nodemuxcore.GetMultiplexer()

	delegator := nodemuxcore.GetDelegatorFactory().GetRESTDelegator(chain.Brand)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	err := delegator.DelegateREST(self.rootCtx, m, chain, method, w, r)
	if err != nil {
		log.Warnf("error delegate rest %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // RESTRelayer.ServeHTTP
