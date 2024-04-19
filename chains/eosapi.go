package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/superisaac/nodemux/core"
	"io"
	"net/http"
)

type eosapiChainInfo struct {
	LastBlockNum int    `json:"last_irreversible_block_num"`
	LastBlockId  string `json:"last_irreversible_block_id"`
}

type eosapiChainGetBlockReq struct {
	BlockNumOrId int `json:"block_num_or_id"`
}

type EOSAPI struct {
}

func NewEOSAPI() *EOSAPI {
	return &EOSAPI{}
}

func (api EOSAPI) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (api EOSAPI) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (api *EOSAPI) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	var chainInfo eosapiChainInfo
	err := ep.PostJson(context,
		"/v1/chain/get_info",
		nil, nil, &chainInfo)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: chainInfo.LastBlockNum,
		Hash:   chainInfo.LastBlockId,
	}
	return block, nil
}

func (api *EOSAPI) DelegateREST(rootCtx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	requiredHeight := -200
	if path == "/v1/chain/get_block" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		getBlockReq := eosapiChainGetBlockReq{}
		if err := json.Unmarshal(body, &getBlockReq); err == nil {
			requiredHeight = getBlockReq.BlockNumOrId
			chain.Log().Infof("retrieved block number %d from get_block request", requiredHeight)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	// } else if path == "/v1/chain/get_infoxx" {
	// 	requiredHeight = 0
	// 	ep, body, err := m.DefaultPipeTeeREST(rootCtx, chain, path, w, r, requiredHeight)
	// 	if err != nil || ep == nil || body == nil {
	// 		return err
	// 	}

	// 	var chainInfo eosapiChainInfo
	// 	if err := json.Unmarshal(body, &chainInfo); err != nil {
	// 		chain.Log().Warnf("json unmarshal body %#v", err)
	// 	}
	// 	block := &nodemuxcore.Block{
	// 		Height: chainInfo.LastBlockNum,
	// 		Hash:   chainInfo.LastBlockId,
	// 	}
	// 	m.UpdateBlockIfChanged(ep, block)
	// 	return nil
	// }
	return m.DefaultPipeREST(rootCtx, chain, path, w, r, requiredHeight)
}
