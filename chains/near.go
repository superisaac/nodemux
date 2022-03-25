package chains

// docsite: https://docs.near.org/docs/api/rpc/

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
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

func (self NearChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self NearChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *NearChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	params := map[string]interface{}{"finality": "final"}
	reqmsg := jsonz.NewRequestMessage(
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

func (self *NearChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
