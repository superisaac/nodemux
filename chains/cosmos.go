package chains

// for cosmos API doc refer to https://v1.cosmos.network/rpc/v0.41.4
// for luna API doc refer to https://lcd.terra.dev/swagger/
// for gRPC and gRPC-Gateway refer to https://buf.build/cosmos/cosmos-sdk/docs/bfe2fb50c22b479e8653f81e23b32659

import (
	"context"
	"fmt"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
)

type cosmosBlock struct {
	BlockID struct {
		Hash string `json:"hash"`
	} `json:"block_id"`

	Block struct {
		Header struct {
			Height      string `json:"height"`
			LastBlockID struct {
				Hash string `json:"hash"`
			} `json:"last_block_id,omitempty"`
		} `json:"header"`
	} `json:"block"`

	height int `json:"-"`
}

func (blk *cosmosBlock) Height() int {
	if blk.height == 0 {
		h, err := strconv.Atoi(blk.Block.Header.Height)
		if err != nil {
			panic(err)
		}
		blk.height = h
	}
	return blk.height
}

type cosmosNodeInfo struct {
	ApplicationVersion struct {
		Name             string `json:"name"`
		AppName          string `json:"app_name"`
		Version          string `json:"version"`
		CosmosSDKVersion string `json:"cosmos_sdk_version"`
	} `json:"application_version"`
}

func (info cosmosNodeInfo) String() string {
	av := info.ApplicationVersion
	return fmt.Sprintf("%s-%s-%s", av.AppName, av.Version, av.CosmosSDKVersion)
}

type CosmosChain struct {
}

func NewCosmosChain() *CosmosChain {
	return &CosmosChain{}
}

func (c CosmosChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	var info cosmosNodeInfo
	err := ep.GetJson(context,
		"/cosmos/base/tendermint/v1beta1/node_info",
		nil, &info)
	if err != nil {
		return "", err
	}
	return info.String(), nil
}

func (c CosmosChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *CosmosChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res cosmosBlock
	err := ep.GetJson(context,
		"/cosmos/base/tendermint/v1beta1/blocks/latest",
		nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.Height(),
		Hash:   res.BlockID.Hash,
	}
	return block, nil
}

func (c *CosmosChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
}
