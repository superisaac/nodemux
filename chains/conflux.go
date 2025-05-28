package chains

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type ConfluxChain struct {
}

func NewConfluxChain() *ConfluxChain {
	return &ConfluxChain{}
}

func (c ConfluxChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c ConfluxChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *ConfluxChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "cfx_epochNumber",
		[]any{"latest_mined"})
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

func (c *ConfluxChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -5)
}
