package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type tronBlock struct {
	BlockHeader struct {
		RawData struct {
			Number     int    `json:"number"`
			ParentHash string `json:"parentHash"`
		} `json:"raw_data"`
	} `json:"block_header"`

	BlockID string `json:"blockID"`
}

type TronChain struct {
}

func NewTronChain() *TronChain {
	return &TronChain{}
}

func (self TronChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *TronChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res tronBlock
	err := ep.PostJson(context,
		"/walletsolidity/getnowblock",
		nil, nil, &res)
	if err != nil {
		return nil, err
	}

	height := res.BlockHeader.RawData.Number
	block := &nodemuxcore.Block{
		Height: height,
		Hash:   res.BlockID,
	}
	return block, nil
}

func (self *TronChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
