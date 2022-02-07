package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func (self *EthereumChain) GetTip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
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

	if ep.Tip == nil || ep.Tip.Height != bt.Height() {
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
	if ep, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg, 0,
		"eth_getTransactionByHash",
		"eth_getTransactionReceipt"); ok {
		return ep.CallRPC(ctx, reqmsg)
	}
	return m.DefaultRelayMessage(ctx, chain, reqmsg, -5)
}
