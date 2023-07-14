package chains

import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type solanaBlock struct {
	Value struct {
		Blockhash            string
		LastValidBlockHeight int `json:"lastValidBlockHeight"`
	}

	Context struct {
		Slot int
	}
}

type SolanaChain struct {
}

func NewSolanaChain() *SolanaChain {
	return &SolanaChain{}
}

func (self SolanaChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self SolanaChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *SolanaChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "getLatestBlockhash", []interface{}{})

	var bt solanaBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: bt.Value.LastValidBlockHeight,
		Hash:   bt.Value.Blockhash,
	}
	return block, nil
}

func (self *SolanaChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -10)
}
