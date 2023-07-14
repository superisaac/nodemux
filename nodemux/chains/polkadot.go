package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type polkadotBlock struct {
	//Hash   string
	Number string

	height int `json:"-"`
}

func (self *polkadotBlock) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
}

type PolkadotChain struct {
}

func NewPolkadotChain() *PolkadotChain {
	return &PolkadotChain{}
}

func (self PolkadotChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self PolkadotChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *PolkadotChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "chain_getHeader", []interface{}{})

	var bt polkadotBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: bt.Height(),
		Hash:   "", //bt.Hash,
	}
	// TODO: get block hash using chain_getBlockHash
	return block, nil
}

func (self *PolkadotChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -2)
}
