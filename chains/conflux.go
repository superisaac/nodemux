package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type ConfluxChain struct {
}

func NewConfluxChain() *ConfluxChain {
	return &ConfluxChain{}
}

func (self *ConfluxChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "cfx_epochNumber",
		[]interface{}{"latest_mined"})
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		var height int
		err := mapstructure.Decode(resmsg.MustResult(), &height)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &nodemuxcore.Block{
			Height: height,
			//Hash:   ""
		}
		return block, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *ConfluxChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
