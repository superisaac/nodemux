package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
)

type tronBlock struct {
	Height int    `mapstructure,"height"`
	Hash   string `mapstructure,"hash"`
}

type TronChain struct {
}

func NewTronChain() *TronChain {
	return &TronChain{}
}

func (self *TronChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	res, err := ep.RequestJson(context,
		"POST",
		"/walletsolidity/getnowblock",
		nil)
	if err != nil {
		return nil, err
	}

	bt := tronBlock{}
	err = mapstructure.Decode(res, &bt)
	if err != nil {
		return nil, errors.Wrap(err, "decode rpcblock")
	}

	block := &balancer.Block{
		Height: bt.Height,
		Hash:   bt.Hash,
	}
	return block, nil
}

func (self *TronChain) RequestREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r)
}
