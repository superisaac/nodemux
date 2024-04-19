package chains

// for substrate API doc refer to https://paritytech.github.io/substrate-api-sidecar/dist/

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
	"strings"
)

type substrateBlock struct {
	Hash   string
	Number string

	height int `json:"-"`
}

func (blk *substrateBlock) Height() int {
	if blk.height == 0 {
		if strings.HasPrefix(blk.Number, "0x") {
			h, err := hexutil.DecodeUint64(blk.Number)
			if err != nil {
				panic(err)
			}
			blk.height = int(h)
		} else {
			h, err := strconv.Atoi(blk.Number)
			if err != nil {
				panic(err)
			}
			blk.height = h
		}
	}
	return blk.height
}

type SubstrateAPI struct {
}

func NewSubstrateAPI() *SubstrateAPI {
	return &SubstrateAPI{}
}

func (c SubstrateAPI) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c SubstrateAPI) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *SubstrateAPI) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res substrateBlock
	err := ep.GetJson(context,
		"/blocks/head",
		nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.Height(),
		Hash:   res.Hash,
	}
	return block, nil
}

func (c *SubstrateAPI) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
}
