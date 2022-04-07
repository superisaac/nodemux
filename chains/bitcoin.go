package chains

// docsite: https://developer.bitcoin.org/reference/rpc/

import (
	"context"
	"fmt"
	"github.com/superisaac/jlib"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"time"
)

type bitcoinBlockhead struct {
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
	reqmsg := jlib.NewRequestMessage(1, "getnetworkinfo", nil)
	var info bitcoinNetworkInfo
	err := ep.UnwrapCallRPC(ctx, reqmsg, &info)
	if err != nil {
		return "", err
	}
	v := fmt.Sprintf("%d", info.Version)
	return v, nil
}

func (self BitcoinChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

// update txid cache from mempool
func (self BitcoinChain) updateMempoolPresenceCache(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) {
	c, ok := m.RedisClient(presenceCacheRedisKey(ep.Chain))
	if !ok {
		return
	}
	reqmsg := jlib.NewRequestMessage(
		1, "getrawmempool", nil)

	var txids []string
	err := ep.UnwrapCallRPC(ctx, reqmsg, &txids)
	if err != nil {
		ep.Log().Warnf("getrawmempool error, %s", err)
		return
	}
	presenceCacheUpdate(
		ctx, c,
		ep.Chain,
		txids,
		ep.Name,
		time.Second*600) // expire after 10 mins
}

func (self BitcoinChain) updateBlockPresenceCache(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint, blockHash string) {
	c, ok := m.RedisClient(presenceCacheRedisKey(ep.Chain))
	if !ok {
		return
	}
	reqmsg := jlib.NewRequestMessage(
		1, "getblock", []interface{}{blockHash})

	var blk bitcoinBlock
	err := ep.UnwrapCallRPC(ctx, reqmsg, &blk)
	if err != nil {
		ep.Log().Warnf("get block error, blockhash %s, %s", blockHash, err)
		return
	}
	presenceCacheUpdate(
		ctx, c,
		ep.Chain,
		blk.Tx,
		ep.Name,
		time.Second*1800) // expire after 30 mins
}

func (self *BitcoinChain) GetBlockhead(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "getchaintips", nil)

	var chaintips []bitcoinBlockhead
	err := ep.UnwrapCallRPC(ctx, reqmsg, &chaintips)
	if err != nil {
		return nil, err
	}
	for _, ct := range chaintips {
		if ct.Status != "active" {
			continue
		}
		block := &nodemuxcore.Block{
			Height: ct.Height,
			Hash:   ct.Hash,
		}

		if ep.Blockhead == nil || ep.Blockhead.Height != ct.Height {
			go self.updateBlockPresenceCache(ctx, m, ep, ct.Hash)
		}
		go self.updateMempoolPresenceCache(ctx, m, ep)
		return block, nil
	}
	return nil, nil
}

func (self *BitcoinChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	if ep, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg,
		"gettransaction",
		"getrawtransaction"); ok {
		return ep.CallRPC(ctx, reqmsg)
	}

	if reqmsg.Method == "getblockhash" {
		if h, ok := self.findBlockHeight(reqmsg); ok {
			return m.DefaultRelayRPC(ctx, chain, reqmsg, h)
		}
	}
	return m.DefaultRelayRPC(ctx, chain, reqmsg, -2)
}

func (self *BitcoinChain) findBlockHeight(reqmsg *jlib.RequestMessage) (int, bool) {
	// the first argument is a integer number
	var bh struct {
		Height int
	}
	if err := jlib.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height > 0 {
		return bh.Height, true
	}
	return 0, false
}
