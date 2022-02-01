package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
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

func (self PolkadotChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *PolkadotChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "chain_getHeader", []interface{}{})
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		bt := polkadotBlock{}
		err := mapstructure.Decode(resmsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &nodemuxcore.Block{
			Height: bt.Number,
			Hash:   bt.Hash,
		}
		return block, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *PolkadotChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -2)
}
