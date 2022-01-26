package chains

import (
	"context"
	"github.com/superisaac/nodemux/multiplex"
	"net/http"
)

type cardanoBlock struct {
	Number int
	Hash   string
}

type cardanoTipResult struct {
	Data struct {
		Blocks []cardanoBlock
	}
}

type cardanoTipRequest struct {
	Query string
}

type CardanoChain struct {
}

func NewCardanoChain() *CardanoChain {
	return &CardanoChain{}
}

func (self *CardanoChain) GetTip(context context.Context, b *multiplex.Multiplexer, ep *multiplex.Endpoint) (*multiplex.Block, error) {
	q := "{blocks(limit:1, order_by:[{number: \"desc\"}]){number hash}}"
	req := cardanoTipRequest{Query: q}
	var res cardanoTipResult
	err := ep.PostJson(context, "", req, nil, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Data.Blocks) > 0 {
		block := &multiplex.Block{
			Height: res.Data.Blocks[0].Number,
			Hash:   res.Data.Blocks[0].Hash,
		}
		return block, nil
	} else {
		ep.Log().Info("query does not return a block")
		return nil, nil
	}

}

func (self *CardanoChain) DelegateGraphQL(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -10)
}
