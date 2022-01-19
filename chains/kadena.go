package chains

import (
	"context"
	"github.com/superisaac/nodemux/nmux"
	"net/http"
)

type kadenaCutBlock struct {
	Height int    `json,"height"`
	Hash   string `json,"hash"`
}

type kadenaCut struct {
	Weight string                    `json,"weight"`
	Hashes map[string]kadenaCutBlock `json,"hashes"`
}

type KadenaChain struct {
}

func NewKadenaChain() *KadenaChain {
	return &KadenaChain{}
}

func (self *KadenaChain) GetTip(context context.Context, b *nmux.Multiplexer, ep *nmux.Endpoint) (*nmux.Block, error) {
	var res kadenaCut
	err := ep.GetJson(context,
		"/chainweb/0.0/mainnet01/cut",
		nil, &res)
	if err != nil {
		return nil, err
	}

	maxHeight := 0
	//maxHash := ""
	for _, block := range res.Hashes {
		if block.Height > maxHeight {
			maxHeight = block.Height
			//maxHash = block.Hash
		}
	}

	block := &nmux.Block{
		Height: maxHeight,
		//Hash: maxHash,
	}
	return block, nil
}

func (self *KadenaChain) DelegateREST(rootCtx context.Context, b *nmux.Multiplexer, chain nmux.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
