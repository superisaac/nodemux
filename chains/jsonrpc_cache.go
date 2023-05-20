package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"strings"
	"time"
)

func jsonrpcCacheGet(ctx context.Context, c *redis.Client, chain nodemuxcore.ChainRef, req *jlib.RequestMessage) (interface{}, bool) {
	cacheKey := req.CacheKey(fmt.Sprintf("CC:%s", chain))
	data, err := c.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}
		log.Warnf("jsonrpcCacheGet(), redis.Get %s: %#v", cacheKey, err)
		return nil, false
	}
	var res interface{}
	resdec := json.NewDecoder(strings.NewReader(data))
	resdec.UseNumber()
	if err := resdec.Decode(&res); err != nil {
		log.Warnf("jsonrpcCacheGet(), parse cache %s: %#v", cacheKey, err)
		return nil, false
	}
	return res, true
}

func jsonrpcCacheUpdate(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, req *jlib.RequestMessage, res *jlib.ResultMessage, expiration time.Duration) {
	if c, ok := m.RedisClientExact(jsonrpcCacheRedisSelector(chain)); ok {
		cacheKey := req.CacheKey(fmt.Sprintf("CC:%s", chain))
		data, err := json.Marshal(res.Result)
		if err != nil {
			log.Warnf("josnrpcCacheUpdate() json.Marshal %s: %#v", cacheKey, err)
			return
		}

		_, err = c.Set(ctx, cacheKey, string(data), expiration).Result()
		if err != nil {
			log.Warnf("josnrpcCacheUpdate() redis.Set %s: %#v", cacheKey, err)
			return
		}
	}
}

func jsonrpcCacheRedisSelector(chain nodemuxcore.ChainRef) string {
	return fmt.Sprintf("jsonrpc-cache-%s-%s", chain.Namespace, chain.Network)
}

func jsonrpcCacheFetchForMethods(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, methods ...string) (bool, *jlib.ResultMessage) {
	useCache := false
	for _, method := range methods {
		if reqmsg.Method == method {
			useCache = true
		}
	}
	if !useCache {
		return false, nil
	}
	if c, ok := m.RedisClientExact(jsonrpcCacheRedisSelector(chain)); ok {
		if resFromCache, ok := jsonrpcCacheGet(ctx, c, chain, reqmsg); ok {
			return useCache, jlib.NewResultMessage(reqmsg, resFromCache)
		} else {
			return useCache, nil
		}
	}
	return false, nil

}
