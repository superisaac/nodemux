package nodemuxcore

import (
	"net/url"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
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

func (self *Multiplexer) RedisClient(selector string) (c *redis.Client, ok bool) {
	if c, ok := self.redisClients[selector]; ok {
		return c, ok
	}

	if store, ok := self.cfg.Stores[selector]; ok && store.Scheme() == "redis" {
		opts, err := GetRedisOptions(store.Url)
		if err != nil {
			log.Panicf("parse redis option error, url=%s, %s", store.Url, err)
			return nil, false
		}
		c := redis.NewClient(opts)
		self.redisClients[selector] = c
		return c, true
	}

	if selector != "default" {
		return self.RedisClient("default")
	} else {
		return nil, false
	}
}

func (self *Multiplexer) RedisClientExact(selector string) (c *redis.Client, ok bool) {
	if c, ok := self.redisClients[selector]; ok {
		return c, ok
	}

	if store, ok := self.cfg.Stores[selector]; ok && store.Scheme() == "redis" {
		opts, err := GetRedisOptions(store.Url)
		if err != nil {
			log.Panicf("parse redis option error, url=%s, %s", store.Url, err)
			return nil, false
		}
		c := redis.NewClient(opts)
		self.redisClients[selector] = c
		return c, true
	}

	return nil, false
}
