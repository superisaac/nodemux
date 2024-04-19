package chains

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type SolanaChain struct {
}

func NewSolanaChain() *SolanaChain {
	return &SolanaChain{}
}

func (c SolanaChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c SolanaChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *SolanaChain) GetBlockhead(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	config := map[string]string{"commitement": "confirmed"}
	reqmsg := jsoff.NewRequestMessage(
		1, "getSlot", []interface{}{config})

	var slot int
	err := ep.UnwrapCallRPC(context, reqmsg, &slot)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: slot,
	}
	return block, nil
}

func (c *SolanaChain) DelegateRPC(rootCtx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return m.DefaultRelayRPC(rootCtx, chain, reqmsg, -10)
}
