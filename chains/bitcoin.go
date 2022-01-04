package chains

import (
	"context"
	//"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/balancer"
)

type BitcoinChaintip struct {
	Status string `mapstructure,"status"`
	Height int    `mapstructure,"height"`
	Hash   string `mapstructure,"hash"`
}

type BitcoinChain struct {
}

func NewBitcoinChain() *BitcoinChain {
	return &BitcoinChain{}
}

func (self *BitcoinChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "getchaintips", []interface{}{})
	resMsg, err := ep.CallHTTP(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		var chaintips []BitcoinChaintip
		err := mapstructure.Decode(resMsg.MustResult(), &chaintips)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		for _, ct := range chaintips {
			if ct.Status != "active" {
				continue
			}
			block := &balancer.Block{
				Height: ct.Height,
				Hash:   ct.Hash,
			}
			return block, nil
		}
		return nil, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *BitcoinChain) RequestReceived(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg)
}
