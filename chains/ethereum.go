package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"time"
	//log "github.com/sirupsen/logrus"
)

type ethereumBlock struct {
	Number       string
	Hash         string
	Transactions []string `json:"transactions"`

	// private fields
	height int
}

func (self *ethereumBlock) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
}

type EthereumChain struct {
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{}
}

func (self *EthereumChain) GetTip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqMsg := jsonz.NewRequestMessage(
		1, "eth_getBlockByNumber",
		[]interface{}{"latest", false})
	resMsg, err := ep.CallRPC(context, reqMsg)
	if err != nil {
		return nil, err
	}
	if resMsg.IsResult() {
		var bt ethereumBlock
		err := mapstructure.Decode(resMsg.MustResult(), &bt)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		block := &nodemuxcore.Block{
			Height: bt.Height(),
			Hash:   bt.Hash,
		}

		if ep.Tip == nil || ep.Tip.Height != bt.Height() {
			go presenceCacheUpdate(context, m, ep.Chain, &bt, ep.Name, time.Second*600) // expire after 10 mins
		}
		return block, nil
	} else {
		errBody := resMsg.MustError()
		return nil, errBody
	}

}

func (self *EthereumChain) DelegateRPC(rootCtx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	if (reqmsg.Method == "eth_getTransactionByHash" ||
		reqmsg.Method == "eth_getTransactionReceipt") &&
		len(reqmsg.Params) > 0 {
		if txHash, ok := reqmsg.Params[0].(string); ok {
			if ep, ok := presenceCacheGetEndpoint(rootCtx, m, chain, txHash); ok {
				resmsg, err := ep.CallRPC(rootCtx, reqmsg)
				return resmsg, err
			}
		}
	}
	return m.DefaultRelayMessage(rootCtx, chain, reqmsg, -5)
}
