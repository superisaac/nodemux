package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/core"
	//"strconv"
)

type polkadotBlock struct {
	Hash   string `mapstructure:"hash"`
	Number int    `mapstructure:"number"`
}

type PolkadotChain struct {
}

func NewPolkadotChain() *PolkadotChain {
	return &PolkadotChain{}
}

func (self *PolkadotChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "chain_getHeader", []interface{}{})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := polkadotBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &nodemuxcore.Block{
			Height: bt.Number,
			Hash:   bt.Hash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *PolkadotChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -2)
}
