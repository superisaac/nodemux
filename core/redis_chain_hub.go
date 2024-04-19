package nodemuxcore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"time"
)

const (
	pubsubKey      = "chain-status"
	snapshotPrefix = "chain-snapshot"
)

// implements ChainStatusHub
type RedisChainhub struct {
	pub  chan ChainStatus
	subs []chan ChainStatus

	rdb *redis.Client

	cmdSub   chan ChCmdChainStatus
	cmdUnsub chan ChCmdChainStatus
}

func NewRedisChainhub(rdb *redis.Client) (*RedisChainhub, error) {
	return &RedisChainhub{
		pub:      make(chan ChainStatus, 100),
		subs:     make([]chan ChainStatus, 0),
		cmdSub:   make(chan ChCmdChainStatus, 10),
		cmdUnsub: make(chan ChCmdChainStatus, 10),
		rdb:      rdb,
	}, nil
}

func (h *RedisChainhub) Sub(ch chan ChainStatus) {
	h.cmdSub <- ChCmdChainStatus{Ch: ch}
}

func (h *RedisChainhub) subscribe(ctx context.Context, ch chan ChainStatus) error {
	h.subs = append(h.subs, ch)
	snKeys, err := h.rdb.Keys(ctx, fmt.Sprintf("%s:*", snapshotPrefix)).Result()
	if err != nil {
		return err
	}
	for _, snKey := range snKeys {
		val, err := h.rdb.Get(ctx, snKey).Result()
		if err != nil {
			return err
		}
		var chainSt ChainStatus
		err = json.Unmarshal([]byte(val), &chainSt)
		if err != nil {
			log.Warnf("error unmarshal %#v", err)
			return err
		}
		ch <- chainSt
	}
	return err
}

func (h *RedisChainhub) Unsub(ch chan ChainStatus) {
	h.cmdUnsub <- ChCmdChainStatus{Ch: ch}
}

func (h *RedisChainhub) unsubscribe(ch chan ChainStatus) {
	found := -1
	for i, sub := range h.subs {
		if sub == ch {
			found = i
			break
		}
	}
	if found >= 0 {
		h.subs = append(h.subs[:found], h.subs[found+1:]...)
	}
}

func (h RedisChainhub) Pub() chan ChainStatus {
	return h.pub
}

func (h *RedisChainhub) listen(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	//rdb := redis.NewClient(h.redisOptions)
	pubsub := h.rdb.Subscribe(ctx, pubsubKey)
	defer pubsub.Close()

	ch := pubsub.Channel(redis.WithChannelSize(1000))

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-ch:
			{
				if !ok {
					log.Warnf("not ok")
					return nil
				}
				var chainSt ChainStatus
				err := json.Unmarshal([]byte(msg.Payload), &chainSt)
				if err != nil {
					log.Warnf("error unmarshal %#v", err)
					return err
				}
				// broadcast to sub channels
				for _, sub := range h.subs {
					sub <- chainSt
				}
			}
		}
	}
}

func (h *RedisChainhub) Run(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	go h.listen(ctx)

	// run publish and other jobs
	networkFailure := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case cmd, ok := <-h.cmdSub:
			if !ok {
				log.Warnf("cmd sub not ok")
				return nil
			}
			err := h.subscribe(ctx, cmd.Ch)
			networkFailure, err = h.handleRedisError(networkFailure, err, "subscribe")
			if err != nil {
				return err
			}
		case cmd, ok := <-h.cmdUnsub:
			if !ok {
				log.Warnf("cmd unsub not ok")
				return nil
			}
			h.unsubscribe(cmd.Ch)
		case chainSt, ok := <-h.pub:
			if !ok {
				log.Warnf("cmd pub not ok")
				return nil
			}

			err := h.publishChainStatus(ctx, chainSt)
			networkFailure, err = h.handleRedisError(networkFailure, err, "publish")
			if err != nil {
				return err
			}
		}
	}
}

func (h *RedisChainhub) publishChainStatus(ctx context.Context, chainSt ChainStatus) error {
	data, err := json.Marshal(chainSt)
	if err != nil {
		return err
	}
	err = h.rdb.Publish(ctx, pubsubKey, data).Err()
	if err != nil {
		return err
	}

	snapshotKey := fmt.Sprintf("%s:%s", snapshotPrefix, chainSt.Chain)
	err = h.rdb.Set(ctx, snapshotKey, data, time.Hour*2).Err()
	if err != nil {
		return err
	}
	return nil
}

func (h RedisChainhub) handleRedisError(networkFailure int, err error, fn string) (int, error) {
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			networkFailure++
			log.Warnf("redis connect failed %d times, %s", networkFailure, opErr)
			if networkFailure > 30 {
				return networkFailure, errors.Wrap(opErr, "networkFailure")
			}
		} else {
			log.Warnf("%s error, %s %s", fn, reflect.TypeOf(err), err)
			return networkFailure, errors.Wrap(err, fn)
		}
	} else {
		networkFailure = 0
	}
	return networkFailure, nil
}
