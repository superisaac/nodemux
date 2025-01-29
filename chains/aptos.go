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

func (b *aptosBlock) Height() int {
	if b.height == 0 {
		h, err := strconv.Atoi(b.BlockHeight)
		if err != nil {
			panic(err)
		}
		b.height = h
	}
	return b.height
}

type AptosAPI struct {
}

func NewAptosAPI() *AptosAPI {
	return &AptosAPI{}
}

func (api AptosAPI) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (api AptosAPI) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (api *AptosAPI) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
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

func (api *AptosAPI) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
}
