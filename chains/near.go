package chains

// docsite: https://docs.near.org/docs/api/rpc/

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type nearBlock struct {
	Header struct {
		Height   int    `json:"height"`
		Hash     string `json:"hash"`
		PrevHash string `json:"prev_hash"`
	} `json:"header"`
}

type NearChain struct {
}

func NewNearChain() *NearChain {
	return &NearChain{}
}

func (c NearChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c NearChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *NearChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	params := map[string]interface{}{"finality": "final"}
	reqmsg := jsoff.NewRequestMessage(
		1, "block", params)

	var bt nearBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Header.Height,
		Hash:   bt.Header.Hash,
	}
	return block, nil
}

func (c *NearChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
