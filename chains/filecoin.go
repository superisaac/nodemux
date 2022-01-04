package chains

import (
	"context"
	//"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/balancer"
)

type FilecoinBlock struct {
	Height int `mapstructure,"Height"`
	//Hash   string `mapstructure,"hash"`
}

type FilecoinChain struct {
}

func NewFilecoinChain() *FilecoinChain {
	return &FilecoinChain{}
}

func (self *FilecoinChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "Filecoin.ChainHead", []interface{}{})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := FilecoinBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		// TODO: get block hash, currently tip.hash is not necessary
		block := &balancer.Block{
			Height: bt.Height,
			//Hash:   bt.Hash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *FilecoinChain) RequestReceived(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg)
}
