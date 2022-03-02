package chains

// Presence cache maintains a map txid -> set[endpoint]. It means
// what endpoints have the txid so that RPC requests can be directed
// to the right node without not found error

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"math/rand"
	"time"
)

func presenceCacheKey(chain nodemuxcore.ChainRef, txid string) string {
	return fmt.Sprintf("P:%s/%s", chain, txid)
}

func presenceCacheUpdate(ctx context.Context, c *redis.Client, chain nodemuxcore.ChainRef, txids []string, epName string, expireAfter time.Duration) {
	for _, txid := range txids {
		key := presenceCacheKey(chain, txid)
		err := c.SAdd(ctx, key, epName).Err()
		if err != nil {
			log.Warnf("error while set %s: %s", key, err)
			return
		}
		err = c.Expire(ctx, key, expireAfter).Err()
		if err != nil {
			log.Warnf("error while expiring key %s: %s", key, err)
			return
		}
	}
}

// try find from healthy endpoint from redis cache
func presenceCacheGetEndpoint(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, txid string) (ep *nodemuxcore.Endpoint, hit bool) {
	c, ok := m.RedisClient(presenceCacheRedisKey(chain))
	if !ok {
		return
	}
	key := presenceCacheKey(chain, txid)
	epNames, err := c.SMembers(ctx, key).Result()
	if err != nil {
		log.Warnf("error getting smembers of %s: %s", key, err)
		return nil, false
	}

	if len(epNames) > 0 {
		// randomly select an endpoint
		epName := epNames[rand.Intn(len(epNames))]
		if ep, ok := m.Get(epName); ok && !ep.Unhealthy {
			return ep, ok
		}

		// sequancially select endpoints
		for _, epName := range epNames {
			if ep, ok := m.Get(epName); ok && !ep.Unhealthy {
				return ep, ok
			}
		}
	}
	return nil, false

}

// match a request message against a given methods list, if matched
// and the firsst param is txid then query the cache for an
// endpoint that has the txid.
func presenceCacheMatchRequest(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage, methods ...string) (*nodemuxcore.Endpoint, bool) {
	found := false
	for _, mth := range methods {
		if reqmsg.Method == mth {
			found = true
			break
		}
	}
	if !found {
		return nil, false
	}

	// struct to extract txid from params
	var txidExtractor struct {
		Txid string
		// Other params are skiped
	}
	err := jsonz.DecodeParams(reqmsg.Params, &txidExtractor)
	if err != nil {
		reqmsg.Log().Warnf("error decoding params for txid: %s", err)
	} else if txidExtractor.Txid != "" {
		return presenceCacheGetEndpoint(ctx, m, chain, txidExtractor.Txid)
	}
	return nil, false
}

func presenceCacheRedisKey(chain nodemuxcore.ChainRef) string {
	return fmt.Sprintf("pcache-%s-%s", chain.Brand, chain.Network)
}
