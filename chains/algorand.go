package chains

import (
	"context"
	//"github.com/mitchellh/mapstructure"
	//"github.com/pkg/errors"
	"github.com/superisaac/nodemux/nmux"
	"net/http"
)

type algorandChainStatus struct {
	LastRound int `json:"lastRound"`
}

type AlgorandChain struct {
}

func NewAlgorandChain() *AlgorandChain {
	return &AlgorandChain{}
}

func (self *AlgorandChain) GetTip(context context.Context, b *nmux.Multiplexer, ep *nmux.Endpoint) (*nmux.Block, error) {
	var res algorandChainStatus
	err := ep.GetJson(context, "/v1/status", nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nmux.Block{
		Height: res.LastRound,
	}
	return block, nil
}

func (self *AlgorandChain) DelegateREST(rootCtx context.Context, b *nmux.Multiplexer, chain nmux.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -5)
}
