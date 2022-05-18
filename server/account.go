package server

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"regexp"
)

var (
	accRegex = regexp.MustCompile(`^/(jsonrpc\-ws|jsonrpc|rest|graphql)/([^/]+)/([^/]+)/([^/]+)`)
)

type Acc struct {
	Name   string
	Chain  nodemuxcore.ChainRef
	Config AccountConfig
}

func NewAccFromConfig(name string, cfg AccountConfig) *Acc {
	return &Acc{
		Name:   name,
		Config: cfg,
	}
}

func AccFromContext(ctx context.Context) *Acc {
	if v := ctx.Value("account"); v != nil {
		if acc, ok := v.(*Acc); ok {
			return acc
		}
		panic("context value account is not an Acc instance")
	}
	panic("context does not have account")
}

type AccHandler struct {
	rootCtx context.Context
	next    http.Handler
}

func NewAccHandler(rootCtx context.Context, next http.Handler) *AccHandler {
	return &AccHandler{
		rootCtx: rootCtx,
		next:    next,
	}
}

func (self *AccHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matches := accRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) < 5 {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return
	}
	account := matches[2]
	brand := matches[3]
	network := matches[4]
	serverCfg := ServerConfigFromContext(self.rootCtx)
	if acccfg, ok := serverCfg.Accounts[account]; ok {
		acc := NewAccFromConfig(account, acccfg)
		acc.Chain = nodemuxcore.ChainRef{
			Brand:   brand,
			Network: network,
		}
		ctx := context.WithValue(r.Context(), "account", acc)
		self.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}
	w.WriteHeader(404)
	w.Write([]byte("not found"))
	return
}
