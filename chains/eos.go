package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
)

type eosChainInfo struct {
	Last_irreversible_block_num int    `mapstructure,"last_irreversible_block_num"`
	Last_irreversible_block_id  string `mapstructure,"last_irreversible_block_id"`
}

type EosChain struct {
}

func NewEosChain() *EosChain {
	return &EosChain{}
}

func (self *EosChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	res, err := ep.RequestJson(context,
		"POST",
		"/v1/chain/get_info",
		nil, nil)
	if err != nil {
		return nil, err
	}

	var tBlock eosChainInfo
	err = mapstructure.Decode(res, &tBlock)
	if err != nil {
		return nil, errors.Wrap(err, "mapst decode eosChainInfo")
	}

	block := &balancer.Block{
		Height: tBlock.Last_irreversible_block_num,
		Hash:   tBlock.Last_irreversible_block_id,
	}
	return block, nil
}

func (self *EosChain) DelegateREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
