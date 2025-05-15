package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"strings"
	"time"
)

func jsonrpcCacheGet(ctx context.Context, m *nodemuxcore.Multiplexer, c *redis.Client, chain nodemuxcore.ChainRef, req *jsoff.RequestMessage) (interface{}, bool) {
	// cacheKey := req.CacheKey(fmt.Sprintf("CC:%s", chain))
	cacheKeys := m.RequestCacheKeys(chain, req, "CC/", -20)
	values, err := c.MGet(ctx, cacheKeys...).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false
		}
		log.Warnf("jsonrpcCacheGet(), redis.Get %s: %#v", cacheKeys, err)
		return nil, false
	}
	for i, value := range values {
		if value == nil {
			continue
		}

		if data, ok := value.(string); ok {
			// return the first cache hits
			var res interface{}
			resdec := json.NewDecoder(strings.NewReader(data))
			resdec.UseNumber()
			if err := resdec.Decode(&res); err != nil {
				log.Warnf("jsonrpcCacheGet(), parse cache %s: %#v", cacheKeys[i], err)
				return nil, false
			}
			return res, true
		}
	}
	return nil, false
}

func jsonrpcCacheUpdate(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint, chain nodemuxcore.ChainRef, req *jsoff.RequestMessage, res *jsoff.ResultMessage, expiration time.Duration) {
	if ep == nil {
		return
	}
	if c, ok := m.RedisClientExact(jsonrpcCacheRedisSelector(chain)); ok {
		cacheKey := req.CacheKey(fmt.Sprintf("CC/%s/", ep.Name))
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

func jsonrpcCacheFetch(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage) (*jsoff.ResultMessage, bool) {
	if c, ok := m.RedisClientExact(jsonrpcCacheRedisSelector(chain)); ok {
		if resFromCache, ok := jsonrpcCacheGet(ctx, m, c, chain, reqmsg); ok {
			return jsoff.NewResultMessage(reqmsg, resFromCache), true
		} else {
			return nil, false
		}
	}
	return nil, false
}

// func jsonrpcCacheFetchForMethods(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, methods ...string) (bool, *jsoff.ResultMessage) {
// 	useCache := false
// 	for _, method := range methods {
// 		if reqmsg.Method == method {
// 			useCache = true
// 		}
// 	}
// 	if !useCache {
// 		return false, nil
// 	}
// 	if c, ok := m.RedisClientExact(jsonrpcCacheRedisSelector(chain)); ok {
// 		if resFromCache, ok := jsonrpcCacheGet(ctx, c, chain, reqmsg); ok {
// 			return useCache, jsoff.NewResultMessage(reqmsg, resFromCache)
// 		} else {
// 			return useCache, nil
// 		}
// 	}
// 	return false, nil

// }
