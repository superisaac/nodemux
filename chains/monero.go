package chains

// docsite: https://monerodocs.org/interacting/monero-wallet-rpc-reference

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type moneroBlock struct {
	Height int
}

type MoneroChain struct {
}

func NewMoneroChain() *MoneroChain {
	return &MoneroChain{}
}

func (c MoneroChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c MoneroChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *MoneroChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
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

func (c *MoneroChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
