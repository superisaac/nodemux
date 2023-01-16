package chains

import (
	"context"
	"fmt"
	"net/http"

	"github.com/superisaac/jlib"
	nodemuxcore "github.com/superisaac/nodemux/core"
)

type txList struct {
	Data []string `json:"data"`
}

type txForTimestamp struct {
	TimestampMs int `json:"timestamp_ms"`
}

type rpcDiscover struct {
	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
}

func (self rpcDiscover) ToString() string {
	return fmt.Sprintf("%s/%s", self.Info.Title, self.Info.Version)
}

type SuiChain struct {
}

func NewSuiChain() *SuiChain {
	return &SuiChain{}
}

func (self SuiChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "rpc.discover", []interface{}{})
	var rpc rpcDiscover
	err := ep.UnwrapCallRPC(context, reqmsg, &rpc)
	if err != nil {
		return "", err
	}

	return rpc.ToString(), nil
}

func (self SuiChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *SuiChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "sui_getTransactions", []interface{}{"All", nil, 2, true})

	var txl txList
	err := ep.UnwrapCallRPC(context, reqmsg, &txl)
	if err != nil {
		return nil, err
	}

	latestDigest := txl.Data[0]

	var tx txForTimestamp

	reqmsg = jlib.NewRequestMessage(
		2, "sui_getTransaction", []interface{}{latestDigest})

	err = ep.UnwrapCallRPC(context, reqmsg, &tx)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: tx.TimestampMs / (1000 * 5), // one block per 5 seconds
		//Hash:   bt.Hash,
	}
	return block, nil
}

func (self *SuiChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
