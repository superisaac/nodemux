package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
)

type eosBlock struct {
	Last_irreversible_block_num int `mapstructure,"last_irreversible_block_num"`
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

	var tBlock eosBlock
	err = mapstructure.Decode(res, &tBlock)
	if err != nil {
		return nil, errors.Wrap(err, "mapst decode eosBlock")
	}

	block := &balancer.Block{
		Height: tBlock.Last_irreversible_block_num,
		//Hash:   tBlock.BlockID,
	}
	return block, nil
}

func (self *EosChain) RequestREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r)
}
