package ratelimit

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	valueBase = 1000
)

type RatelimitOptions struct {
	Time time.Time
	Span time.Duration
}

func Incr(context context.Context, c *redis.Client, field string, limit int, optslist ...*RatelimitOptions) (ok bool, e error) {
	span1h, _ := time.ParseDuration("1h")
	opts := &RatelimitOptions{
		Time: time.Now(),
		Span: span1h,
	}
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

	spanSecs := int(opts.Span.Seconds())
	if spanSecs <= 0 {
		panic("negative time span")
	}

	var ts int = int(opts.Time.Unix()) / spanSecs
	key := fmt.Sprintf("rtlm:%d:%d", spanSecs, ts)
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
		if err := c.ExpireNX(context, key, expiration).Err(); err != nil {
			return false, err
		}

		return true, nil
	} else {
		return true, nil
	}
}
