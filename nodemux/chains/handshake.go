package chains

// docsite: https://hsd-dev.org/api-docs/

import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type handshakeBlockhead struct {
	Status string
	Height int
	Hash   string
}

type handshakeBlock struct {
	Hash   string
	Height int
	Tx     []string
}

// see hsd-cli getinfo
type handshakeInfo struct {
	Version string
}

type HandshakeChain struct {
}

func NewHandshakeChain() *HandshakeChain {
	return &HandshakeChain{}
}

func (self HandshakeChain) GetClientVersion(ctx context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jlib.NewRequestMessage(1, "getinfo", nil)
	var info handshakeInfo
	err := ep.UnwrapCallRPC(ctx, reqmsg, &info)
	if err != nil {
		return "", err
	}
	return info.Version, nil
}

func (self HandshakeChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *HandshakeChain) GetBlockhead(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "getchaintips", nil)

	var chaintips []handshakeBlockhead
	err := ep.UnwrapCallRPC(ctx, reqmsg, &chaintips)
	if err != nil {
		return nil, err
	}
	for _, ct := range chaintips {
		if ct.Status != "active" {
			continue
		}
		block := &nodemuxcore.Block{
			Height: ct.Height,
			Hash:   ct.Hash,
		}
		return block, nil
	}
	return nil, nil
}

func (self *HandshakeChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	return m.DefaultRelayRPC(ctx, chain, reqmsg, -2)
}
