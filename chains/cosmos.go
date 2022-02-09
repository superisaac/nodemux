package chains

// for cosmos API doc refer to https://v1.cosmos.network/rpc/v0.41.4
// for luna API doc refer to https://lcd.terra.dev/swagger/

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
)

type cosmosBlock struct {
	Block struct {
		Header struct {
			Height string
		}
	}

	height int
}

func (self *cosmosBlock) Height() int {
	if self.height == 0 {
		h, err := strconv.Atoi(self.Block.Header.Height)
		if err != nil {
			panic(err)
		}
		self.height = h
	}
	return self.height
}

type CosmosChain struct {
}

func NewCosmosChain() *CosmosChain {
	return &CosmosChain{}
}

func (self CosmosChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *CosmosChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res cosmosBlock
	err := ep.GetJson(context,
		"/blocks/latest",
		nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.Height(),
		// Hash: not provided
	}
	return block, nil
}

func (self *CosmosChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
}
