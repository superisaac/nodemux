package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type kadenaCutBlock struct {
	Height int
	Hash   string
}

type kadenaCut struct {
	Weight string
	Hashes map[string]kadenaCutBlock
}

type KadenaChain struct {
}

func NewKadenaChain() *KadenaChain {
	return &KadenaChain{}
}

func (self KadenaChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self KadenaChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *KadenaChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
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

	block := &nodemuxcore.Block{
		Height: maxHeight,
		//Hash: maxHash,
	}
	return block, nil
}

func (self *KadenaChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -30)
}
