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
		Username: u.User.Username(),
		Password: pwd,
		DB:       db,
	}
	return opt, nil
}

func (m *Multiplexer) RedisClient(selector string) (c *redis.Client, ok bool) {
	if c, ok := m.redisClients[selector]; ok {
		return c, ok
	}

	if store, ok := m.cfg.Stores[selector]; ok && store.Scheme() == "redis" {
		opts, err := GetRedisOptions(store.Url)
		if err != nil {
			log.Panicf("parse redis option error, url=%s, %s", store.Url, err)
			return nil, false
		}
		c := redis.NewClient(opts)
		m.redisClients[selector] = c
		return c, true
	}

	if selector != "default" {
		return m.RedisClient("default")
	} else {
		return nil, false
	}
}

func (m *Multiplexer) RedisClientExact(selector string) (c *redis.Client, ok bool) {
	if c, ok := m.redisClients[selector]; ok {
		return c, ok
	}

	if store, ok := m.cfg.Stores[selector]; ok && store.Scheme() == "redis" {
		opts, err := GetRedisOptions(store.Url)
		if err != nil {
			log.Panicf("parse redis option error, url=%s, %s", store.Url, err)
			return nil, false
		}
		c := redis.NewClient(opts)
		m.redisClients[selector] = c
		return c, true
	}

	return nil, false
}
