package chains

// docsite: https://nippycodes.com/coding/casper-node-openrpc/

import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type casperBlock struct {
	Header struct {
		Height int
	}
	Hash string
}

type CasperChain struct {
}

func NewCasperChain() *CasperChain {
	return &CasperChain{}
}

func (self CasperChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self CasperChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *CasperChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "chain_get_block", nil)

	var bt casperBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Header.Height,
		Hash:   bt.Hash,
	}
	return block, nil
}

func (self *CasperChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
