package server

import (
	"context"
	"net/http"

	"github.com/superisaac/nodemux/core"
	"github.com/superisaac/nodemux/ratelimit"
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

func (handler *RatelimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	acc := AccFromContextOrNil(r.Context())
	if acc == nil {
		handler.next.ServeHTTP(w, r)
		return
	}
	var ratelimit RatelimitConfig
	var accName string
	if acc != nil {
		ratelimit = acc.Config.Ratelimit
		accName = acc.Config.Username
		if accName == "" {
			accName = acc.Name
		}
	} else {
		serverCfg := ServerConfigFromContext(handler.rootCtx)
		ratelimit = serverCfg.Ratelimit
		accName = ""
	}

	ok, err := checkRatelimit(r, accName, ratelimit, false)
	if err != nil {
		requestLog(r).Errorf("error while checking ratelimit %s", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write([]byte(`{"error": {"code": 500, "messasge": "server error"}, "id": null}`))

		// w.WriteHeader(500)
		// w.Write([]byte("server error"))
	} else if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		w.Write([]byte(`{"error": {"code": 429, "messasge": "ratelimit exceeded!"}, "id": null}`))
		// acceptableMediaTypes := []contenttype.MediaType{
		// 	contenttype.NewMediaType("application/json"),
		// }

		// if _, _, err := contenttype.GetAcceptableMediaType(r, acceptableMediaTypes); err == nil {
		// 	w.Header().Set("Content-Type", "application/json")
		// 	w.WriteHeader(429)
		// 	w.Write([]byte(`{"error": {"code": 429, "messasge": "ratelimit exceeded!"}, "id": null}`))
		// } else {
		// 	w.Header().Set("Content-Type", "text/plain")
		// 	w.WriteHeader(429)
		// 	w.Write([]byte("rate limit exceeded!"))
		// }
	} else {
		handler.next.ServeHTTP(w, r)
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
