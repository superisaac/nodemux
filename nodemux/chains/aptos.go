package chains

// for aptos API doc refer to https://fullnode.mainnet.aptoslabs.com/v1/spec

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
)

type aptosBlock struct {
	BlockHeight string `json:"block_height"`

	height int `json:"-"`
}

func (self *aptosBlock) Height() int {
	if self.height == 0 {
		h, err := strconv.Atoi(self.BlockHeight)
		if err != nil {
			panic(err)
		}
		self.height = h
	}
	return self.height
}

type AptosAPI struct {
}

func NewAptosAPI() *AptosAPI {
	return &AptosAPI{}
}

func (self AptosAPI) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self AptosAPI) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *AptosAPI) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res aptosBlock
	err := ep.GetJson(context,
		"",
		nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.Height(),
		Hash:   "",
	}
	return block, nil
}

func (self *AptosAPI) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
}
