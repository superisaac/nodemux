package chains

import (
	"context"
	"fmt"
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
	if resMsg.IsRequest() {
		bt := RPCBlock{}
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}
		block := &balancer.Block{
			Height:   1, // TODO: web3 convert hex to int
			Hash:     bt.Hash,
			PrevHash: bt.PrevHash,
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errors.New(fmt.Sprintf("Get tip error %d %s", errBody.Code, errBody.Message))
	}

}
