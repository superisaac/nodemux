package chains

import (
	"context"
	//"github.com/mitchellh/mapstructure"
	//"github.com/pkg/errors"
	"github.com/superisaac/nodepool/balancer"
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

func (self *EosChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	var res eosChainInfo
	err := ep.RequestJson(context,
		"POST",
		"/v1/chain/get_info",
		nil, nil, &res)
	if err != nil {
		return nil, err
	}

	block := &balancer.Block{
		Height: res.LastBlockNum,
		Hash:   res.LastBlockId,
	}
	return block, nil
}

func (self *EosChain) DelegateREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
