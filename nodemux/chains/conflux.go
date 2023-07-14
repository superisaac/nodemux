package chains

import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type ConfluxChain struct {
}

func NewConfluxChain() *ConfluxChain {
	return &ConfluxChain{}
}

func (self ConfluxChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self ConfluxChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *ConfluxChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "cfx_epochNumber",
		[]interface{}{"latest_mined"})
	var height int
	err := ep.UnwrapCallRPC(context, reqmsg, &height)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: height,
		//Hash:   ""
	}
	return block, nil
}

func (self *ConfluxChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -5)
}
