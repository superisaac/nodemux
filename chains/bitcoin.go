package chains

import (
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/nodemux/core"
	"time"
)

type bitcoinChaintip struct {
	Status string
	Height int
	Hash   string
}

type bitcoinBlock struct {
	Hash   string
	Height int
	Tx     []string
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

func (self BitcoinChain) GetClientVersion(ctx context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsonz.NewRequestMessage(1, "getnetworkinfo", nil)
	resmsg, err := ep.CallRPC(ctx, reqmsg)
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

func (self *BitcoinChain) updatePresenceCache(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint, blockHash string) {
	reqmsg := jsonz.NewRequestMessage(
		1, "getblock", []interface{}{blockHash})

	resmsg, err := ep.CallRPC(ctx, reqmsg)
	if err != nil {
		ep.Log().Warnf("get block error, blockhash %s, %s", blockHash, err)
		return
	}
	if resmsg.IsResult() {
		var blk bitcoinBlock
		err := mapstructure.Decode(resmsg.MustResult(), &blk)
		if err != nil {
			ep.Log().Warnf("decode block error, blockhash %s, %s", blockHash, err)
			return
		}
		go presenceCacheUpdate(
			ctx, m,
			ep.Chain,
			blk.Tx, ep.Name,
			time.Second*1800) // expire after 30 mins
	} else {
		ep.Log().Warnf("get block error msg, blockhash %s, %s", blockHash, resmsg.MustError())
	}

}

func (self *BitcoinChain) GetTip(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "getchaintips", nil)
	resmsg, err := ep.CallRPC(ctx, reqmsg)
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

			if ep.Tip == nil || ep.Tip.Height != ct.Height {
				go self.updatePresenceCache(ctx, m, ep, ct.Hash)
			}
			return block, nil
		}
		return nil, nil
	} else {
		errBody := resmsg.MustError()
		return nil, errBody
	}

}

func (self *BitcoinChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	if ep, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg, 0,
		"gettransaction",
		"getrawtransaction"); ok {
		return ep.CallRPC(ctx, reqmsg)
	}
	return m.DefaultRelayMessage(ctx, chain, reqmsg, -2)
}
