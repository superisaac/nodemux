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

func (self Multiplexer) RedisClient() (c *redis.Client, ok bool) {
	if self.cfg.Store.Scheme() != "redis" {
		return nil, false
	}
	if self.redisClient == nil {
		opts, err := GetRedisOptions(self.cfg.Store.Url)
		if err != nil {
			log.Panicf("parse redis option error, %s", err)
		}
		self.redisClient = redis.NewClient(opts)
	}
	return self.redisClient, true
}
