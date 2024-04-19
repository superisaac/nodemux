package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type polkadotBlock struct {
	//Hash   string
	Number string

	height int `json:"-"`
}

func (blk *polkadotBlock) Height() int {
	if blk.height <= 0 {
		height, err := hexutil.DecodeUint64(blk.Number)
		if err != nil {
			panic(err)
		}
		blk.height = int(height)
	}
	return blk.height
}

type PolkadotChain struct {
}

func NewPolkadotChain() *PolkadotChain {
	return &PolkadotChain{}
}

func (c PolkadotChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c PolkadotChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *PolkadotChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "chain_getHeader", []interface{}{})

	var bt polkadotBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: bt.Height(),
		Hash:   "", //bt.Hash,
	}
	// TODO: get block hash using chain_getBlockHash
	return block, nil
}

func (c *PolkadotChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -2)
}
