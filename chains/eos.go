package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type eosChainInfo struct {
	LastBlockNum int    `json:"last_irreversible_block_num"`
	LastBlockId  string `json:"last_irreversible_block_id"`
}

type EosChain struct {
}

func NewEosChain() *EosChain {
	return &EosChain{}
}

func (self EosChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *EosChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res eosChainInfo
	err := ep.PostJson(context,
		"/v1/chain/get_info",
		nil, nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.LastBlockNum,
		Hash:   res.LastBlockId,
	}
	return block, nil
}

func (self *EosChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
