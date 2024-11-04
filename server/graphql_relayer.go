package server

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

// GraphQL Handler
type GraphQLRelayer struct {
	rootCtx context.Context
	//regex   *regexp.Regexp
	//chain   nodemuxcore.ChainRef
	acc *Acc
}

func NewGraphQLRelayer(rootCtx context.Context) *GraphQLRelayer {
	return &GraphQLRelayer{
		rootCtx: rootCtx,
		//regex:   regexp.MustCompile(`^/graphql/([^/]+\/[^/]+)(/.*)?$`),
	}
}

func (h *GraphQLRelayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acc := h.acc
	path := "/"
	if acc == nil {
		acc = AccFromContext(r.Context())
		if acc == nil {
			w.WriteHeader(404)
			w.Write([]byte("acc not found"))
			return
		}
	}

	m := nodemuxcore.GetMultiplexer()
	delegator := nodemuxcore.GetDelegatorFactory().GetGraphQLDelegator(acc.Chain.Namespace)
	if delegator == nil {
		w.WriteHeader(404)
		w.Write([]byte("backend not found"))
		return
	}

	err := delegator.DelegateGraphQL(h.rootCtx, m, acc.Chain, path, w, r)
	if err != nil {
		requestLog(r).Warnf("error delegate graphql %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}
} // GraphQLRelayer.ServeHTTP
