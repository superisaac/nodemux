package chains

import (
	"context"
	//"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/balancer"
)

type solanaBlockValue struct {
	Blockhash            string `mapstructure,"blockhash"`
	LastValidBlockHeight int    `mapstructure,"lastValidBlockHeight"`
}
type solanaBlockContext struct {
	Slot int `mapstructure,"slot"`
}
type solanaBlock struct {
	Value   solanaBlockValue   `mapstructure,"value"`
	Context solanaBlockContext `mapstructure,"context"`
}

type SolanaChain struct {
}

func NewSolanaChain() *SolanaChain {
	return &SolanaChain{}
}

func (self *SolanaChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "getLatestBlockhash", []interface{}{})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := solanaBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		block := &balancer.Block{
			Height: bt.Value.LastValidBlockHeight,
			Hash:   bt.Value.Blockhash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *SolanaChain) RequestReceived(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg)
}
