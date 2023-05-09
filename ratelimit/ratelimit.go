package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	valueBase = 1000
)

type RatelimitOptions struct {
	Time time.Time
	Span time.Duration
}

func NewRatelimitOptions() *RatelimitOptions {
	span1h, _ := time.ParseDuration("1h")
	opts := &RatelimitOptions{
		Time: time.Now(),
		Span: span1h,
	}
	return opts
}

func (opts *RatelimitOptions) RedisKey() string {
	spanSecs := int(opts.Span.Seconds())
	if spanSecs <= 0 {
		panic("negative time span")
	}

	var ts int = int(opts.Time.Unix()) / spanSecs
	key := fmt.Sprintf("rtlm:%d:%d", spanSecs, ts)
	return key
}

func Incr(context context.Context, c *redis.Client, field string, limit int, optslist ...*RatelimitOptions) (ok bool, e error) {
	opts := NewRatelimitOptions()
	for _, srcopt := range optslist {
		if srcopt == nil {
			continue
		}
		// copy the opt content
		if !srcopt.Time.IsZero() {
			opts.Time = srcopt.Time
		}
		if srcopt.Span != 0 {
			opts.Span = srcopt.Span
		}
	}

	key := opts.RedisKey()
	i64value, err := c.HIncrBy(context, key, field, 1).Result()
	if err != nil {
		return false, err
	}

	newValue := int(i64value)
	if newValue >= valueBase+limit {
		// out of limit, return false
		return false, nil
	} else if newValue <= valueBase {
		// field has not being set previously
		if err := c.HSet(context, key, field, valueBase).Err(); err != nil {
			return false, err
		}
		expiration := opts.Span * 2
		// ExpireNX only works abover version redis 7.0.0
		//if err := c.ExpireNX(context, key, expiration).Err(); err != nil {
		//	return false, err
		//}

		if err := c.Expire(context, key, expiration).Err(); err != nil {
			return false, err
		}

		return true, nil
	} else {
		return true, nil
	}
}

func Values(context context.Context, c *redis.Client, optslist ...*RatelimitOptions) (map[string]int64, error) {
	opts := NewRatelimitOptions()
	for _, srcopt := range optslist {
		if srcopt == nil {
			continue
		}
		// copy the opt content
		if !srcopt.Time.IsZero() {
			opts.Time = srcopt.Time
		}
		if srcopt.Span != 0 {
			opts.Span = srcopt.Span
		}
	}

	key := opts.RedisKey()
	items, err := c.HGetAll(context, key).Result()
	if err != nil {
		return nil, err
	}
	used := make(map[string]int64)
	for field, value := range items {
		i, err := strconv.ParseInt(value, 10, 64)
		if err == nil && i >= valueBase {
			used[field] = i - valueBase
		}
	}
	return used, nil
}
