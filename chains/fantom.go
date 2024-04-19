package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type fantomBlock struct {
	Number string
	Hash   string

	height int `json:"-"`
}

func (blk *fantomBlock) Height() int {
	if blk.height <= 0 {
		height, err := hexutil.DecodeUint64(blk.Number)
		if err != nil {
			panic(err)
		}
		blk.height = int(height)
	}
	return blk.height
}

type fantomTipResult struct {
	Block fantomBlock
}

type FantomChain struct {
}

func NewFantomChain() *FantomChain {
	return &FantomChain{}
}

func (c FantomChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c FantomChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *FantomChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	q := "{block(){number hash}}"
	var res fantomTipResult
	err := ep.RequestGraphQL(context, q, nil, nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.Block.Height(),
		Hash:   res.Block.Hash,
	}
	return block, nil
}

func (c *FantomChain) DelegateGraphQL(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -10)
}
