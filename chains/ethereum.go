package chains

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"time"
)

type ethereumBlock struct {
	Number       string
	Hash         string
	Transactions []string `json:"transactions"`

	// private fields
	height int
}

func (self *ethereumBlock) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
}

type EthereumChain struct {
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{}
}

func (self EthereumChain) cacheKey(chain nodemuxcore.ChainRef, txHash string) string {
	return fmt.Sprintf("Q:%s/%s", chain, txHash)
}

func (self *EthereumChain) updateTxCache(m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, block *ethereumBlock, epName string) {
	c, ok := m.RedisClient()
	if !ok {
		// no redis connection
		return
	}
	ctx := context.Background()
	for _, txHash := range block.Transactions {
		key := self.cacheKey(chain, txHash)
		err := c.SAdd(ctx, key, epName).Err()
		if err != nil {
			log.Warnf("error while set %s: %s", key, err)
			return
		}
		err = c.Expire(ctx, key, time.Second*600).Err() // expire after 10 mins
		if err != nil {
			log.Warnf("error while expiring key %s: %s", key, err)
			return
		}
	}
}
func (self *EthereumChain) endpointFromCache(m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, txHash string) (ep *nodemuxcore.Endpoint, hit bool) {
	key := self.cacheKey(chain, txHash)
	c, ok := m.RedisClient()
	if !ok {
		// no redis connection
		return
	}
	epNames, err := c.SMembers(context.Background(), key).Result()
	if err != nil {
		log.Warnf("error getting smembers of %s: %s", key, err)
		return nil, false
	}

	for _, epName := range epNames {
		if ep, ok := m.Get(epName); ok && ep.Healthy {
			return ep, ok
		}
	}
	return nil, false

}

func (self *EthereumChain) GetTip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqMsg := jsonz.NewRequestMessage(
		1, "eth_getBlockByNumber",
		[]interface{}{"latest", false})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		var bt ethereumBlock
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &nodemuxcore.Block{
			Height: bt.Height(),
			Hash:   bt.Hash,
		}

		if ep.Tip == nil || ep.Tip.Height != bt.Height() {
			go self.updateTxCache(m, ep.Chain, &bt, ep.Name)
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *EthereumChain) DelegateRPC(rootCtx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	if (reqmsg.Method == "eth_getTransactionByHash" ||
		reqmsg.Method == "eth_getTransactionReceipt") &&
		len(reqmsg.Params) > 0 {
		if txHash, ok := reqmsg.Params[0].(string); ok {
			if ep, ok := self.endpointFromCache(m, chain, txHash); ok {
				resmsg, err := ep.CallRPC(rootCtx, reqmsg)
				return resmsg, err
			}
		}
	}
	return m.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
