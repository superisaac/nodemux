package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"strconv"
)

type starcoinBlock struct {
	Head struct {
		Blockhash string `mapstructure:"block_hash"`
		Number    string `mapstructure:"number"`
	} `mapstructure:"head"`
}

type StarcoinChain struct {
}

func NewStarcoinChain() *StarcoinChain {
	return &StarcoinChain{}
}

func (self *StarcoinChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "chain.info", nil)
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		bt := starcoinBlock{}
		err := mapstructure.Decode(resmsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		height, err := strconv.Atoi(bt.Head.Number)
		if err != nil {
			return nil, errors.Wrap(err, "strconv.Atoi")
		}

		block := &nodemuxcore.Block{
			Height: height,
			Hash:   bt.Head.Blockhash,
		}
		return block, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *StarcoinChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -3)
}
