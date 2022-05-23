package chains

// for cosmos API doc refer to https://v1.cosmos.network/rpc/v0.41.4
// for luna API doc refer to https://lcd.terra.dev/swagger/

import (
	"context"
	"fmt"
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

type cosmosNodeInfo struct {
	ApplicationVersion struct {
		ServerName       string `json:"server_name"`
		Version          string `json:"version"`
		CosmosSDKVersion string `json:"cosmos_sdk_version"`
	} `json:"application_version"`
}

func (self cosmosNodeInfo) String() string {
	av := self.ApplicationVersion
	return fmt.Sprintf("%s-%s-%s", av.ServerName, av.Version, av.CosmosSDKVersion)
}

type CosmosChain struct {
}

func NewCosmosChain() *CosmosChain {
	return &CosmosChain{}
}

func (self CosmosChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	var info cosmosNodeInfo
	err := ep.GetJson(context, "/node_info", nil, &info)
	if err != nil {
		return "", err
	}
	return info.String(), nil
}

func (self CosmosChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *CosmosChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
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
