package chains

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/balancer"
)

type rippleLedger struct {
	LedgerIndex int `mapstructure:"ledger_index`
}

type RippleChain struct {
}

func NewRippleChain() *RippleChain {
	return &RippleChain{}
}

func (self *RippleChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	filter := map[string]interface{}{
		"ledger_index": "validated",
		"accounts":     false,
		"full":         false,
		"transactions": false,
		"expand":       false,
		"owner_funds":  false,
	}
	reqMsg := jsonrpc.NewRequestMessage(
		1, "ledger", []interface{}{filter})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	fmt.Printf("ledger %#v\n", resMsg)
	if resMsg.IsResult() {
		var ledger rippleLedger
		err := mapstructure.Decode(resMsg.MustResult(), &ledger)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &balancer.Block{
			Height: ledger.LedgerIndex,
			//Hash:   ct.Hash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *RippleChain) DelegateRPC(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -10)
}
