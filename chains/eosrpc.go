package chains

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type eosrpcChainInfo struct {
	LastBlockNum int    `json:"last_irreversible_block_num"`
	LastBlockId  string `json:"last_irreversible_block_id"`
}

// type eosrpcChainGetBlockReq struct {
// 	BlockNumOrId int `json:"block_num_or_id"`
// }

type EOSRPC struct {
}

func NewEOSRPC() *EOSRPC {
	return &EOSRPC{}
}

func (c EOSRPC) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c EOSRPC) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *EOSRPC) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "get_info", nil)

	var info eosrpcChainInfo
	err := ep.UnwrapCallRPC(context, reqmsg, &info)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: info.LastBlockNum,
		Hash:   info.LastBlockId,
	}
	return block, nil
}

func (c *EOSRPC) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -300)
}
