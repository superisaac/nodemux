package chains

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/balancer"
)

type RPCBlock struct {
	Number   string `mapstructure,"number"`
	Hash     string `mapstructure,"hash"`
	PrevHash string `mapstructure,"prevhash"`
}

type EthereumChain struct {
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{}
}

func (self *EthereumChain) GetTip(context context.Context, ep *balancer.Endpoint) (*balancer.Block, error) {
	reqMsg := jsonrpc.NewRequestMessage(
		1, "eth_getBlockByNumber",
		[]interface{}{"latest", false})
	resMsg, err := ep.CallHTTP(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		bt := RPCBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		height, err := hexutil.DecodeUint64(bt.Number)
		if err != nil {
			return nil, errors.Wrapf(err, "hexutil.decode %s", bt.Number)
		}
		block := &balancer.Block{
			Height:   int(height),
			Hash:     bt.Hash,
			PrevHash: bt.PrevHash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errors.New(fmt.Sprintf("Get tip error %d %s", errBody.Code, errBody.Message))
	}

}
