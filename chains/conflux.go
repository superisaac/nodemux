package chains

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type ConfluxChain struct {
}

func NewConfluxChain() *ConfluxChain {
	return &ConfluxChain{}
}

func (self ConfluxChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *ConfluxChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "cfx_epochNumber",
		[]interface{}{"latest_mined"})
	var height int
	err := ep.UnwrapCallRPC(context, reqmsg, &height)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: height,
		//Hash:   ""
	}
	return block, nil
}

func (self *ConfluxChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -5)
}
