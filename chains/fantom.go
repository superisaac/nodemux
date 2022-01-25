package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/nodemux/multiplex"
	"net/http"
)

type fantomBlock struct {
	Number string
	Hash   string

	height int
}

func (self *fantomBlock) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
}

type fantomTipResult struct {
	Data struct {
		Block fantomBlock
	}
}

type fantomTipRequest struct {
	Query string
}

type FantomChain struct {
}

func NewFantomChain() *FantomChain {
	return &FantomChain{}
}

func (self *FantomChain) GetTip(context context.Context, b *multiplex.Multiplexer, ep *multiplex.Endpoint) (*multiplex.Block, error) {
	q := "{block(){number hash}}"
	req := fantomTipRequest{Query: q}
	var res fantomTipResult
	err := ep.PostJson(context, "", req, nil, &res)
	if err != nil {
		return nil, err
	}

	block := &multiplex.Block{
		Height: res.Data.Block.Height(),
		Hash:   res.Data.Block.Hash,
	}
	return block, nil
}

func (self *FantomChain) DelegateGraphQL(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -10)
}
