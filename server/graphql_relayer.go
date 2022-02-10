package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

// GraphQL Handler
type GraphQLRelayer struct {
	rootCtx context.Context
	regex   *regexp.Regexp
	chain   nodemuxcore.ChainRef
}

func NewGraphQLRelayer(rootCtx context.Context) *GraphQLRelayer {
	return &GraphQLRelayer{
		rootCtx: rootCtx,
		regex:   regexp.MustCompile(`^/graphql/([^/]+)/([^/]+)(/.*)?$`),
	}
}

func (self *GraphQLRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chain := self.chain
	path := "/"
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
		path = matches[3]
		chain = nodemuxcore.ChainRef{
			Brand:   brand,
			Network: network,
		}
	}

	m := nodemuxcore.GetMultiplexer()
	delegator := nodemuxcore.GetDelegatorFactory().GetGraphQLDelegator(chain.Brand)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	err := delegator.DelegateGraphQL(self.rootCtx, m, chain, path, w, r)
	if err != nil {
		log.Warnf("error delegate rest %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // GraphQLRelayer.ServeHTTP
