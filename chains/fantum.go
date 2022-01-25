package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/nodemux/multiplex"
	"net/http"
)

type fantumBlock struct {
	Number string
	Hash   string

	height int
}

func (self *fantumBlock) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
}

type fantumTipResult struct {
	Data struct {
		Block fantumBlock
	}
}

type fantumTipRequest struct {
	Query string
}

type FantumChain struct {
}

func NewFantumChain() *FantumChain {
	return &FantumChain{}
}

func (self *FantumChain) GetTip(context context.Context, b *multiplex.Multiplexer, ep *multiplex.Endpoint) (*multiplex.Block, error) {
	q := "{block(){number hash}}"
	req := fantumTipRequest{Query: q}
	var res fantumTipResult
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

func (self *FantumChain) DelegateGraphQL(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, w, r, -10)
}
