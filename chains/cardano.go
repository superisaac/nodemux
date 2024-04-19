package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type cardanoBlock struct {
	Number int
	Hash   string
}

type cardanoTipResult struct {
	Blocks []cardanoBlock
}

type CardanoChain struct {
}

func NewCardanoChain() *CardanoChain {
	return &CardanoChain{}
}

func (c CardanoChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c CardanoChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *CardanoChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	q := "{blocks(limit:1, order_by:[{number: \"desc\"}]){number hash}}"
	var res cardanoTipResult
	err := ep.RequestGraphQL(context, q, nil, nil, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Blocks) > 0 {
		block := &nodemuxcore.Block{
			Height: res.Blocks[0].Number,
			Hash:   res.Blocks[0].Hash,
		}
		return block, nil
	} else {
		ep.Log().Info("query does not return a block")
		return nil, nil
	}

}

func (c *CardanoChain) DelegateGraphQL(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -10)
}
