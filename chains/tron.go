package chains

import (
	"context"
	//"github.com/pkg/errors"
	"github.com/superisaac/nodemux/nmux"
	"net/http"
	//"strconv"
)

type tronBlockRawData struct {
	Number     int    `json:"number"`
	ParentHash string `json:"parentHash"`
}

type tronBlockHeader struct {
	RawData tronBlockRawData `json:"raw_data"`
}

type tronBlock struct {
	BlockHeader tronBlockHeader `json:"block_header"`
	BlockID     string          `json:"blockID"`
}

type TronChain struct {
}

func NewTronChain() *TronChain {
	return &TronChain{}
}

func (self *TronChain) GetTip(context context.Context, b *nmux.Multiplexer, ep *nmux.Endpoint) (*nmux.Block, error) {
	var res tronBlock
	err := ep.PostJson(context,
		"/walletsolidity/getnowblock",
		nil, nil, &res)
	if err != nil {
		return nil, err
	}

	height := res.BlockHeader.RawData.Number
	block := &nmux.Block{
		Height: height,
		Hash:   res.BlockID,
	}
	return block, nil
}

func (self *TronChain) DelegateREST(rootCtx context.Context, b *nmux.Multiplexer, chain nmux.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
