package chains

import (
	"context"
	"github.com/pkg/errors"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
)

type starcoinBlock struct {
	Head struct {
		Blockhash string `json:"block_hash"`
		Number    string `json:"number"`
	} `json:"head"`
}

type StarcoinChain struct {
}

func NewStarcoinChain() *StarcoinChain {
	return &StarcoinChain{}
}

func (c StarcoinChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c StarcoinChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *StarcoinChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "chain.info", nil)

	var bt starcoinBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}
	height, err := strconv.Atoi(bt.Head.Number)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}

	block := &nodemuxcore.Block{
		Height: height,
		Hash:   bt.Head.Blockhash,
	}
	return block, nil
}

func (c *StarcoinChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
