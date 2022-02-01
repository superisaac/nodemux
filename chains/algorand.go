package chains

import (
	"context"
	//"github.com/mitchellh/mapstructure"
	//"github.com/pkg/errors"
	"github.com/superisaac/nodemux/core"
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

func (self AlgorandChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *AlgorandChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res algorandChainStatus
	err := ep.GetJson(context, "/v1/status", nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.LastRound,
	}
	return block, nil
}

func (self *AlgorandChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -5)
}
