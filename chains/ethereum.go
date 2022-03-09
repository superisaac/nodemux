package chains

// docsite: https://ethereum.org/en/developers/docs/apis/json-rpc/

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
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

type ethereumStreaming struct {
	subscribeError error
	currentBlock   *nodemuxcore.Block
}

type ethereumHeadSub struct {
	Subscription string
	Result       ethereumBlock
}

type EthereumChain struct {
	streamings map[string]*ethereumStreaming
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{
		streamings: make(map[string]*ethereumStreaming),
	}
}

func (self EthereumChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "web3_clientVersion", nil)
	var v string
	err := ep.UnwrapCallRPC(context, reqmsg, &v)
	if err != nil {
		ep.Log().Warnf("error call rpc web3_clientVersion %s", err)
		return "", err
	}
	return v, nil
}

func (self *EthereumChain) GetChaintip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	if blk, ok := self.fetchStreaming(context, ep); ok && blk != nil {
		return blk, nil
	}

	reqmsg := jsonz.NewRequestMessage(
		jsonz.NewUuid(), "eth_getBlockByNumber",
		[]interface{}{"latest", false})

	var bt ethereumBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Height(),
		Hash:   bt.Hash,
	}

	if ep.Chaintip == nil || ep.Chaintip.Height != bt.Height() {
		if c, ok := m.RedisClient(presenceCacheRedisKey(ep.Chain)); ok {
			go presenceCacheUpdate(
				context, c,
				ep.Chain,
				bt.Transactions, ep.Name,
				time.Second*600) // expire after 10 mins
		}
	}
	return block, nil
}

func (self *EthereumChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	if endpoint, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg,
		"eth_getTransactionByHash",
		"eth_getTransactionReceipt"); ok {
		return endpoint.CallRPC(ctx, reqmsg)
	}

	if reqmsg.Method == "eth_getBlockByNumber" {
		if h, ok := self.findBlockHeight(reqmsg); ok {
			return m.DefaultRelayRPC(ctx, chain, reqmsg, h)
		}
	}
	return m.DefaultRelayRPC(ctx, chain, reqmsg, -5)
}

func (self *EthereumChain) findBlockHeight(reqmsg *jsonz.RequestMessage) (int, bool) {
	// the first argument is a hexlified block number or latest or pending
	var bh struct {
		Height string
	}
	if err := jsonz.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height != "" {
		if bh.Height == "latest" || bh.Height == "pending" {
			return 0, true
		}
		if height, err := hexutil.DecodeUint64(bh.Height); err == nil {
			return int(height), true
		}
	}
	return 0, false
}

func (self *EthereumChain) fetchStreaming(context context.Context, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, bool) {
	if !ep.IsWebsocket() {
		return nil, false
	}

	if es, ok := self.streamings[ep.Name]; ok {
		if es.subscribeError != nil {
			return nil, false
		}
		return es.currentBlock, true
	}

	es := &ethereumStreaming{}
	self.subscribeChaintip(context, ep, es)
	return nil, false
}

func (self *EthereumChain) subscribeChaintip(rootCtx context.Context, ep *nodemuxcore.Endpoint, es *ethereumStreaming) {
	if wsClient, ok := ep.RPCClient().(*jsonzhttp.WSClient); ok {
		var subscribeToken string
		self.streamings[ep.Name] = es

		submsg := jsonz.NewRequestMessage(
			jsonz.NewUuid(), "eth_subscribe",
			[]interface{}{"newHeads"})

		err := ep.UnwrapCallRPC(rootCtx, submsg, &subscribeToken)
		if err != nil {
			//panic(err)
			ep.Log().Warnf("subscribe error %s", err)
			es.subscribeError = err
			return
		}
		ep.Log().Debugf("eth got subscrib subscribeToken %s", subscribeToken)

		// listening eth_subscription notify
		wsClient.OnMessage(func(msg jsonz.Message) {
			ntf, ok := msg.(*jsonz.NotifyMessage)
			if !ok && ntf.Method != "eth_subscription" || len(ntf.Params) == 0 {
				return
			}
			var headSub ethereumHeadSub
			err := jsonz.DecodeInterface(ntf.Params[0], &headSub)
			if err != nil {
				ep.Log().Warnf("decode head sub error %s", err)
			} else {
				// match Subscription against sub token
				if headSub.Subscription != subscribeToken {
					ep.Log().Warnf("subscription %s != subscribeToken %s", headSub.Subscription, subscribeToken)
					return
				}
				es.currentBlock = &nodemuxcore.Block{
					Height: headSub.Result.Height(),
					Hash:   headSub.Result.Hash,
				}
			}
		})
	}
}
