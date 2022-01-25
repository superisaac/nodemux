package chains

import (
	"context"
	"github.com/superisaac/nodemux/multiplex"
	"net/http"
)

type rippleLedgerResult struct {
	Result struct {
		LedgerIndex int `mapstructure:"ledger_index json:"ledger_index"`
	}
}

type rippleLedgerFilter struct {
	LedgerIndex  string `json:"ledger_index"`
	Accounts     bool
	Full         bool
	Transactions bool
	Expand       bool
	OwnerFunds   bool `json: "owner_funds"`
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

func (self *RippleChain) GetTip(context context.Context, b *multiplex.Multiplexer, ep *multiplex.Endpoint) (*multiplex.Block, error) {
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
	block := &multiplex.Block{
		Height: res.Result.LedgerIndex,
		//Hash:   ct.Hash,
	}
	return block, nil

	// reqMsg := jsonrpc.NewRequestMessage(
	// 	1, "ledger", []interface{}{filter})
	// resMsg, err := ep.CallRPC(context, reqMsg)
	// if err != nil {
	// 	return nil, err
	// }
	// if resMsg.IsResult() {
	// 	var ledger rippleLedger
	// 	err := mapstructure.Decode(resMsg.MustResult(), &ledger)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "decode rpcblock")
	// 	}

	// 	block := &multiplex.Block{
	// 		Height: ledger.LedgerIndex,
	// 		//Hash:   ct.Hash,
	// 	}
	// 	return block, nil
	// } else {
	// 	errBody := resMsg.MustError()
	// 	return nil, errBody
	// }

}

func (self *RippleChain) DelegateREST(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r, -10)
}

// func (self *RippleChain) DelegateREST(rootCtx context.Context, b *multiplex.Multiplexer, chain multiplex.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
// 	// Custom relay methods can be defined here
// 	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -10)
// }
