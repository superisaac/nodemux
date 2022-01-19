package chains

import (
	"context"
	//"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodemux/nmux"
)

type EthereumBlock struct {
	Number string `mapstructure:"number"`
	Hash   string `mapstructure:"hash"`
}

type EthereumChain struct {
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{}
}

func (self *EthereumChain) GetTip(context context.Context, b *nmux.Multiplexer, ep *nmux.Endpoint) (*nmux.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "eth_getBlockByNumber",
		[]interface{}{"latest", false})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := EthereumBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		height, err := hexutil.DecodeUint64(bt.Number)
		if err != nil {
			return nil, errors.Wrapf(err, "hexutil.decode %s", bt.Number)
		}
		block := &nmux.Block{
			Height: int(height),
			Hash:   bt.Hash,
			//PrevHash: bt.ParentHash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *EthereumChain) DelegateRPC(rootCtx context.Context, b *nmux.Multiplexer, chain nmux.ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
