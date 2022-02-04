package chains

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
)

type bitcoinChaintip struct {
	Status string `mapstructure:"status"`
	Height int    `mapstructure:"height"`
	Hash   string `mapstructure:"hash"`
}

// see bitcoin-cli getnetworkinfo
type bitcoinNetworkInfo struct {
	Version int
}

type BitcoinChain struct {
}

func NewBitcoinChain() *BitcoinChain {
	return &BitcoinChain{}
}

func (self BitcoinChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsonz.NewRequestMessage(1, "getnetworkinfo", nil)
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return "", err
	}
	if resmsg.IsResult() {
		var info bitcoinNetworkInfo
		err := mapstructure.Decode(resmsg.MustResult(), &info)
		if err != nil {
			return "", errors.Wrap(err, "decode network info")
		} else {
			v := fmt.Sprintf("%d", info.Version)
			return v, nil
		}
	} else {
		return "", resmsg.MustError()
	}
}

func (self *BitcoinChain) GetTip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "getchaintips", nil)
	resmsg, err := ep.CallRPC(context, reqmsg)
	if err != nil {
		return nil, err
	}
	if resmsg.IsResult() {
		var chaintips []bitcoinChaintip
		err := mapstructure.Decode(resmsg.MustResult(), &chaintips)
		if err != nil {
			return nil, errors.Wrap(err, "decode rpcblock")
		}

		for _, ct := range chaintips {
			if ct.Status != "active" {
				continue
			}
			block := &nodemuxcore.Block{
				Height: ct.Height,
				Hash:   ct.Hash,
			}
			return block, nil
		}
		return nil, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *BitcoinChain) DelegateRPC(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	// Custom relay methods can be defined here
	return b.DefaultRelayMessage(rootCtx, chain, reqmsg, -2)
}
