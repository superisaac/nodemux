package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	valueBase = int64(1000)
)

func IncrHourly(context context.Context, c *redis.Client, field string, limit int64) (ok bool, e error) {
	span, _ := time.ParseDuration("1m")
	return incr(context, c, field, limit, "hour", span)
}

func IncrMinutely(context context.Context, c *redis.Client, field string, limit int64) (ok bool, e error) {
	span, _ := time.ParseDuration("1m")
	return incr(context, c, field, limit, "min", span)
}

func incr(context context.Context, c *redis.Client, field string, limit int64, prefix string, timeSpan time.Duration) (ok bool, e error) {
	var ts int64 = (time.Now().Unix()) / int64(timeSpan.Seconds())
	key := fmt.Sprintf("rtlm:%s:%d", prefix, ts)
	newValue, err := c.HIncrBy(context, key, field, 1).Result()
	if err != nil {
		return false, err
	}
	if newValue >= valueBase+limit {
		// out of limit, return false
		return false, nil
	} else if newValue <= valueBase {
		// field has not being set previously
		if err := c.HSet(context, key, field, valueBase).Err(); err != nil {
			return false, err
		}
		expiration := timeSpan * 2
		if err := c.ExpireNX(context, key, expiration).Err(); err != nil {
			return false, err
		}

		return true, nil
	} else {
		return true, nil
	}
}
