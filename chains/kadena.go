package chains

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
)

type kadenaCutBlock struct {
	Height int    `mapstructure,"height"`
	Hash   string `mapstructure,"hash"`
}

type kadenaCut struct {
	Weight string                    `mapstructure,"weight"`
	Hashes map[string]kadenaCutBlock `mapstructure,"hashes"`
}

type KadenaChain struct {
}

func NewKadenaChain() *KadenaChain {
	return &KadenaChain{}
}

func (self *KadenaChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	res, err := ep.GetJson(context,
		"/chainweb/0.0/mainnet01/cut",
		nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("res %#v\n", res)

	var tBlock kadenaCut
	err = mapstructure.Decode(res, &tBlock)
	if err != nil {
		return nil, errors.Wrap(err, "mapst decode kadenaChainInfo")
	}

	maxHeight := 0
	//maxHash := ""
	for _, block := range tBlock.Hashes {
		if block.Height > maxHeight {
			maxHeight = block.Height
			//maxHash = block.Hash
		}
	}

	block := &balancer.Block{
		Height: maxHeight,
		//Hash: maxHash,
	}
	return block, nil
}

func (self *KadenaChain) DelegateREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
