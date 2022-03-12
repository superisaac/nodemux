package chains

// docsite: https://monerodocs.org/interacting/monero-wallet-rpc-reference

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type moneroBlock struct {
	Height int
}

type MoneroChain struct {
}

func NewMoneroChain() *MoneroChain {
	return &MoneroChain{}
}

func (self MoneroChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self MoneroChain) StartFetch(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *MoneroChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "get_height", nil)

	var bt moneroBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Height,
		//Hash:   bt.Hash,
	}
	return block, nil
}

func (self *MoneroChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
