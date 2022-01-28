package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
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
	reqMsg := jsonrpc.NewRequestMessage(
		1, "chain.info", nil)
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := starcoinBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
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
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *StarcoinChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -3)
}
