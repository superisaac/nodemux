package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type rippleLedgerResult struct {
	Result struct {
		LedgerIndex int `json:"ledger_index"`
	}
}

type rippleLedgerFilter struct {
	LedgerIndex  string `json:"ledger_index"`
	Accounts     bool
	Full         bool
	Transactions bool
	Expand       bool
	OwnerFunds   bool `json:"owner_funds"`
}

type rippleLedgerRequest struct {
	Method string
	Params []rippleLedgerFilter
}

type RippleChain struct {
}

func NewRippleChain() *RippleChain {
	return &RippleChain{}
}

func (self RippleChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self RippleChain) StartFetch(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (self *RippleChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	filter := rippleLedgerFilter{
		LedgerIndex:  "validated",
		Accounts:     false,
		Full:         false,
		Transactions: false,
		Expand:       false,
		OwnerFunds:   false,
	}
	req := rippleLedgerRequest{
		Method: "ledger",
		Params: []rippleLedgerFilter{filter},
	}
	var res rippleLedgerResult
	err := ep.PostJson(context, "", req, nil, &res)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: res.Result.LedgerIndex,
		//Hash:   ct.Hash,
	}
	return block, nil

}

func (self *RippleChain) DelegateREST(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -10)
}
