package server

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"github.com/superisaac/nodemux/ratelimit"
	"net/http"
)

// handle ratelimit
type RatelimitHandler struct {
	rootCtx context.Context
	next    http.Handler
}

func NewRatelimitHandler(rootCtx context.Context, next http.Handler) *RatelimitHandler {
	return &RatelimitHandler{
		rootCtx: rootCtx,
		next:    next,
	}
}

func (self *RatelimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acc := AccFromContext(r.Context())
	var ratelimit RatelimitConfig
	var accName string
	if acc != nil {
		ratelimit = acc.Config.Ratelimit
		accName = acc.Config.Username
		if accName == "" {
			accName = acc.Name
		}
	} else {
		serverCfg := ServerConfigFromContext(self.rootCtx)
		ratelimit = serverCfg.Ratelimit
		accName = ""
	}

	ok, err := checkRatelimit(r, accName, ratelimit, false)
	if err != nil {
		requestLog(r).Errorf("error while checking ratelimit %s", err)
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	} else if !ok {
		w.WriteHeader(429)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("rate limit exceeded!"))
	} else {
		self.next.ServeHTTP(w, r)
	}
}

func checkRatelimit(r *http.Request, accountName string, ratelimitCfg RatelimitConfig, fromWebsocket bool) (bool, error) {
	m := nodemuxcore.GetMultiplexer()
	factor := 1
	if fromWebsocket {
		factor = 2
	}
	if c, ok := m.RedisClient("ratelimit"); ok {
		if accountName != "" {
			// use account based limit
			return ratelimit.Incr(
				r.Context(),
				c,
				//"u"+accName,
				accountName,
				ratelimitCfg.UserLimit()*factor)
		} else {
			// per IP based ratelimit
			return ratelimit.Incr(
				r.Context(),
				c, r.RemoteAddr,
				ratelimitCfg.IPLimit()*factor)
		}
	}
	return true, nil
}
