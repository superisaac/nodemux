package chains

import (
	"context"
	"fmt"
	"net/http"

	"github.com/superisaac/jsoff"
	nodemuxcore "github.com/superisaac/nodemux/core"
)

type txInfo struct {
	Digest      string `json:"digest"`
	TimestampMs int    `json:"timestampMs"`
}

type txList struct {
	Data []txInfo `json:"data"`
}

type txQuery struct {
	Options struct {
		ShowRawInput bool `json:"showRawInput"`
	} `json:"options"`
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
	reqmsg := jsoff.NewRequestMessage(
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
	query := &txQuery{}
	query.Options.ShowRawInput = true

	reqmsg := jsoff.NewRequestMessage(
		1, "suix_queryTransactionBlocks",
		[]interface{}{query, nil, 2, true})

	var txl txList
	err := ep.UnwrapCallRPC(context, reqmsg, &txl)
	if err != nil {
		return nil, err
	}

	if len(txl.Data) <= 0 {
		return nil, nil
	}

	latestTx := txl.Data[0]

	block := &nodemuxcore.Block{
		Height: latestTx.TimestampMs / (1000 * 5), // one dummy block per 5 seconds
		//Hash:   bt.Hash,
	}
	return block, nil
}

func (self *SuiChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
