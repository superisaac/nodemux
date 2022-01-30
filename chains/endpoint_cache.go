package chains

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/core"
	"math/rand"
	"time"
)

func txCacheKey(chain nodemuxcore.ChainRef, txHash string) string {
	return fmt.Sprintf("Q:%s/%s", chain, txHash)
}

func updateTxCache(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, block *ethereumBlock, epName string, expireAfter time.Duration) {
	c, ok := m.RedisClient()
	if !ok {
		// no redis connection
		return
	}
	for _, txHash := range block.Transactions {
		key := txCacheKey(chain, txHash)
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
func endpointFromTxCache(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, txHash string) (ep *nodemuxcore.Endpoint, hit bool) {
	key := txCacheKey(chain, txHash)
	c, ok := m.RedisClient()
	if !ok {
		// no redis connection
		return nil, false
	}
	epNames, err := c.SMembers(ctx, key).Result()
	if err != nil {
		log.Warnf("error getting smembers of %s: %s", key, err)
		return nil, false
	}

	if len(epNames) > 0 {
		// randomly select an endpoint
		epName := epNames[rand.Intn(len(epNames))]
		if ep, ok := m.Get(epName); ok && ep.Healthy {
			return ep, ok
		}

		// sequancially select endpoints
		for _, epName := range epNames {
			if ep, ok := m.Get(epName); ok && ep.Healthy {
				return ep, ok
			}
		}
	}
	return nil, false

}
