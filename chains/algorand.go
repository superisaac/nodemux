package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
)

type algorandBlock struct {
	LastRound int `mapstructure,"lastRound"`
}

type AlgorandChain struct {
}

func NewAlgorandChain() *AlgorandChain {
	return &AlgorandChain{}
}

func (self *AlgorandChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	res, err := ep.GetJson(context, "/v1/status", nil)
	if err != nil {
		return nil, err
	}

	var tBlock algorandBlock
	err = mapstructure.Decode(res, &tBlock)
	if err != nil {
		return nil, errors.Wrap(err, "mapst decode algorandBlock")
	}

	block := &balancer.Block{
		Height: tBlock.LastRound,
		//Hash:   tBlock.Last_irreversible_block_id,
	}
	return block, nil
}

func (self *AlgorandChain) DelegateREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r)
}
