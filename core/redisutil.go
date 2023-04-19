package nodemuxcore

import (
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strconv"
)

func GetRedisOptions(redisUrl string) (*redis.Options, error) {
	u, err := url.Parse(redisUrl)
	if err != nil {
		return nil, err
	}
	sdb := u.Path[1:]
	db := 0
	if sdb != "" {
		db, err = strconv.Atoi(sdb)
		if err != nil {
			return nil, err
		}
	}
	pwd, ok := u.User.Password()
	if !ok {
		pwd = ""
	}

	opt := &redis.Options{
		Addr:     u.Host,
		Password: pwd,
		DB:       db,
	}
	return opt, nil
}

func (self *Multiplexer) RedisClient(key string) (c *redis.Client, ok bool) {
	if c, ok := self.redisClients[key]; ok {
		return c, ok
	}

	if store, ok := self.cfg.Stores[key]; ok && store.Scheme() == "redis" {
		opts, err := GetRedisOptions(store.Url)
		if err != nil {
			log.Panicf("parse redis option error, url=%s, %s", store.Url, err)
			return nil, false
		}
		c := redis.NewClient(opts)
		self.redisClients[key] = c
		return c, true
	}

	if key != "default" {
		return self.RedisClient("default")
	} else {
		return nil, false
	}
}
