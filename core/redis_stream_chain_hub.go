package nodemuxcore

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"
	"time"
)

const (
	streamsKey = "chain-status-streams"
)

// implements ChainStatusHub
type RedisStreamChainhub struct {
	pub  chan ChainStatus
	subs []chan ChainStatus

	rdb *redis.Client

	cmdSub   chan ChCmdChainStatus
	cmdUnsub chan ChCmdChainStatus
}

func NewRedisStreamChainhub(rdb *redis.Client) (*RedisStreamChainhub, error) {
	return &RedisStreamChainhub{
		pub:      make(chan ChainStatus, 100),
		subs:     make([]chan ChainStatus, 0),
		cmdSub:   make(chan ChCmdChainStatus, 10),
		cmdUnsub: make(chan ChCmdChainStatus, 10),
		rdb:      rdb,
	}, nil
}

func (h *RedisStreamChainhub) Sub(ch chan ChainStatus) {
	h.cmdSub <- ChCmdChainStatus{Ch: ch}
}

func (h *RedisStreamChainhub) subscribe(ctx context.Context, ch chan ChainStatus) error {
	h.subs = append(h.subs, ch)

	// got the last 100 items
	revmsgs, err := h.rdb.XRevRangeN(ctx, streamsKey, "+", "-", 100).Result()
	if err != nil {
		return errors.Wrap(err, "redis.XRevRangeN(100)")
	}

	sent := make(map[string]bool)
	for _, xmsg := range revmsgs {
		chainSt, err := h.decodeChainStatus(&xmsg)
		if err != nil {
			return err
		}
		if _, ok := sent[chainSt.EndpointName]; !ok {
			sent[chainSt.EndpointName] = true
			ch <- chainSt
		}
	}
	return nil
}

func (h RedisStreamChainhub) decodeChainStatus(xmsg *redis.XMessage) (ChainStatus, error) {
	val, ok := xmsg.Values["cst"]
	if !ok {
		return ChainStatus{}, errors.New("stream item has no cst")
	}
	sval, ok := val.(string)
	if !ok {
		return ChainStatus{}, errors.New("value of cst is not string")
	}
	var chainSt ChainStatus
	err := json.Unmarshal([]byte(sval), &chainSt)
	if err != nil {
		return ChainStatus{}, err
	}
	return chainSt, nil
}

func (h *RedisStreamChainhub) Unsub(ch chan ChainStatus) {
	h.cmdUnsub <- ChCmdChainStatus{Ch: ch}
}

func (h *RedisStreamChainhub) unsubscribe(ch chan ChainStatus) {
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

func (h RedisStreamChainhub) Pub() chan ChainStatus {
	return h.pub
}

func (h *RedisStreamChainhub) listen(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	var lastID string = ""

	for {
		var xmsgs []redis.XMessage
		var err error
		if lastID == "" {
			// get the last item
			xmsgs, err = h.rdb.XRevRangeN(ctx, streamsKey, "+", "-", 1).Result()
		} else {
			// open start by prefixing "("
			xmsgs, err = h.rdb.XRangeN(ctx, streamsKey, "("+lastID, "+", 10).Result()
		}
		if err != nil {
			return errors.Wrap(err, "read range n")
		}
		log.Debugf("refdis stream got %d msgs, %s", len(xmsgs), lastID)
		if len(xmsgs) <= 0 {
			// sleep for a while when no new streaming data
			time.Sleep(time.Millisecond * 3)
		} else {
			for _, xmsg := range xmsgs {
				chainSt, err := h.decodeChainStatus(&xmsg)
				if err != nil {
					return err
				}
				lastID = xmsg.ID
				// broadcast to subscribe channels
				for _, sub := range h.subs {
					sub <- chainSt
				}

			}
		}
	}
}

func (h *RedisStreamChainhub) Run(rootCtx context.Context) error {
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

func (h *RedisStreamChainhub) publishChainStatus(ctx context.Context, chainSt ChainStatus) error {

	data, err := json.Marshal(chainSt)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	values := map[string]interface{}{"cst": string(data)}
	//fmt.Printf("values ssss %s\n", values)

	id, err := h.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamsKey,
		Values: values,
		MaxLen: 1000,
	}).Result()

	if err != nil {
		return errors.Wrap(err, "redis.XAdd")
	}

	log.Debugf("xadd got id %s", id)
	return nil
}

func (h RedisStreamChainhub) handleRedisError(networkFailure int, err error, fn string) (int, error) {
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
