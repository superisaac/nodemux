package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/superisaac/nodemux/core"
	"io"
	"net/http"
)

type eosChainInfo struct {
	LastBlockNum int    `json:"last_irreversible_block_num"`
	LastBlockId  string `json:"last_irreversible_block_id"`
}

type eosChainGetBlockReq struct {
	BlockNumOrId int `json:"block_num_or_id"`
}

type EosChain struct {
}

func NewEosChain() *EosChain {
	return &EosChain{}
}

func (self EosChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self EosChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *EosChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var res eosChainInfo
	err := ep.PostJson(context,
		"/v1/chain/get_info",
		nil, nil, &res)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: res.LastBlockNum,
		Hash:   res.LastBlockId,
	}
	return block, nil
}

func (self *EosChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	h := -10
	if path == "/v1/chain/get_block" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		getBlockReq := eosChainGetBlockReq{}
		if err := json.Unmarshal(body, &getBlockReq); err == nil {
			h = getBlockReq.BlockNumOrId
			chain.Log().Infof("retrieved block number %d from get_block request", h)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	} else if path == "/v1/chain/get_info" {
		h = 0
	}
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, h)
}
