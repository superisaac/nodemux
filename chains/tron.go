package chains

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
	"strconv"
)

type tronBlockRawData struct {
	Number     string `mapstructure,"number"`
	ParentHash string `mapstructure,"parentHash"`
}

type tronBlockHeader struct {
	Raw_data tronBlockRawData `mapstructure,"raw_data"`
}

type tronBlock struct {
	Block_header tronBlockHeader `mapstructure,"block_header"`
	BlockID      string          `mapstructure,"blockID"`
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
		nil, nil)
	if err != nil {
		return nil, err
	}

	var tBlock tronBlock
	err = mapstructure.Decode(res, &tBlock)
	if err != nil {
		return nil, errors.Wrap(err, "mapst decode tronBlock")
	}

	height, err := strconv.Atoi(tBlock.Block_header.Raw_data.Number)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}

	block := &balancer.Block{
		Height: height,
		Hash:   tBlock.BlockID,
	}
	return block, nil
}

func (self *TronChain) RequestREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r)
}
