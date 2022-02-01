package chains

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"strings"
	"time"
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

func (self EthereumChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqId := strings.ReplaceAll(uuid.New().String(), "-", "")
	reqmsg := jsonz.NewRequestMessage(
		reqId, "web3_clientVersion",
		[]interface{}{})
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		ep.Log().Warnf("error call rpc web3_clientVersion %s", err)
		return "", err
	}
	if resmsg.IsResult() {
		var v string
		err := mapstructure.Decode(resmsg.MustResult(), &v)
		if err != nil {
			return "", errors.Wrap(err, "decode client version")
		} else {
			return v, nil
		}
	} else {
		return "", resmsg.MustError()
	}

}

func (self *EthereumChain) GetTip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqId := strings.ReplaceAll(uuid.New().String(), "-", "")
	reqmsg := jsonz.NewRequestMessage(
		reqId, "eth_getBlockByNumber",
		[]interface{}{"latest", false})
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		var bt ethereumBlock
		err := mapstructure.Decode(resmsg.MustResult(), &bt)
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
		errBody := resmsg.MustError()
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
