package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hashicorp/golang-lru"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
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

type ethereumTxIndex struct {
	lruCache *lru.Cache
}

type EthereumChain struct {
	txIndexes map[nodemuxcore.ChainRef](*ethereumTxIndex)
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{
		txIndexes: make(map[nodemuxcore.ChainRef]*ethereumTxIndex),
	}
}

func (self *EthereumChain) updateTxCache(chain nodemuxcore.ChainRef, block *ethereumBlock, epName string) {
	idx, ok := self.txIndexes[chain]
	if !ok {
		cache, err := lru.New(1024)
		if err != nil {
			panic(err)
		}
		idx = &ethereumTxIndex{
			lruCache: cache,
		}
		self.txIndexes[chain] = idx
	}
	for _, txHash := range block.Transactions {
		idx.lruCache.Add(txHash, epName)
	}
}

func (self *EthereumChain) endpointFromCache(chain nodemuxcore.ChainRef, b *nodemuxcore.Multiplexer, txHash string) (ep *nodemuxcore.Endpoint, hit bool) {
	if idx, ok := self.txIndexes[chain]; ok {
		if v, ok := idx.lruCache.Get(txHash); ok {
			epName, ok := v.(string)
			if !ok {
				log.Panicf("epName for txHash %s is not string", txHash)
				return nil, false
			}

			if ep, ok := b.Get(epName); ok && ep.Healthy {
				return ep, ok
			}
		}
	}
	return nil, false

}

func (self *EthereumChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
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
			go self.updateTxCache(ep.Chain, &bt, ep.Name)
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *EthereumChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	if (reqmsg.Method == "eth_getTransactionByHash" ||
		reqmsg.Method == "eth_getTransactionReceipt") &&
		len(reqmsg.Params) > 0 {
		if txHash, ok := reqmsg.Params[0].(string); ok {
			if ep, ok := self.endpointFromCache(chain, b, txHash); ok {
				resmsg, err := ep.CallRPC(rootCtx, reqmsg)
				return resmsg, err
			}
		}
	}
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
