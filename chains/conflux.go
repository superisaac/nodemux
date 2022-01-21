package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/multiplex"
)

type ConfluxChain struct {
}

func NewConfluxChain() *ConfluxChain {
	return &ConfluxChain{}
}

func (self *ConfluxChain) GetTip(context context.Context, b *multiplex.Multiplexer, ep *multiplex.Endpoint) (*multiplex.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "cfx_epochNumber",
		[]interface{}{"latest_mined"})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		var height int
		err := mapstructure.Decode(resMsg.MustResult(), &height)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &multiplex.Block{
			Height: height,
			//Hash:   ""
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *ConfluxChain) DelegateRPC(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
