package chains

import (
	"context"
	//"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type solanaBlock struct {
	Value struct {
		Blockhash            string `mapstructure:"blockhash"`
		LastValidBlockHeight int    `mapstructure:"lastValidBlockHeight"`
	} `mapstructure:"value"`

	Context struct {
		Slot int `mapstructure:"slot"`
	} `mapstructure:"context"`
}

type SolanaChain struct {
}

func NewSolanaChain() *SolanaChain {
	return &SolanaChain{}
}

func (self *SolanaChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "getLatestBlockhash", []interface{}{})
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		bt := solanaBlock{}
		err := mapstructure.Decode(resmsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		block := &nodemuxcore.Block{
			Height: bt.Value.LastValidBlockHeight,
			Hash:   bt.Value.Blockhash,
		}
		return block, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *SolanaChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -10)
}
