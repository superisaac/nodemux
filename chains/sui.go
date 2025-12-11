package chains

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/superisaac/jsoff"
	nodemuxcore "github.com/superisaac/nodemux/core"
	"net/http"
	"strconv"
)

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

func (d rpcDiscover) ToString() string {
	return fmt.Sprintf("%s/%s", d.Info.Title, d.Info.Version)
}

type SuiChain struct {
}

func NewSuiChain() *SuiChain {
	return &SuiChain{}
}

func (c SuiChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "rpc.discover", []any{})
	var rpc rpcDiscover
	err := ep.UnwrapCallRPC(context, reqmsg, &rpc)
	if err != nil {
		return "", err
	}

	return rpc.ToString(), nil
}

func (c SuiChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *SuiChain) GetBlockhead(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	query := &txQuery{}
	query.Options.ShowRawInput = true

	reqmsg := jsoff.NewRequestMessage(
		1, "sui_getLatestCheckpointSequenceNumber",
		[]any{query, nil, 2, true})

	var seqString string
	err := ep.UnwrapCallRPC(context, reqmsg, &seqString)
	if err != nil {
		return nil, errors.Wrap(err, "call.latestCheckpoint")
	}

	seq, err := strconv.Atoi(seqString)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}

	block := &nodemuxcore.Block{
		Height: seq,
		//Hash:   bt.Hash,
	}
	return block, nil
}

func (c *SuiChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayRPC(rootCtx, chain, reqmsg, -3)
}
