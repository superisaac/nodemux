package chains

// docsite: https://developer.bitcoin.org/reference/rpc/

import (
	"context"
	"fmt"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"time"
)

type bitcoinBlockchainInfo struct {
	Blocks        int    `json:"blocks"`
	BestBlockhash string `json:"bestblockhash"`
}

// type bitcoinBlock struct {
// 	Hash   string
// 	Height int
// 	Tx     []string
// }

// see bitcoin-cli getnetworkinfo
type bitcoinNetworkInfo struct {
	Version         int    `json:"version"`
	SubVersion      string `json:"subversion"`
	ProtocolVersion int    `json:"protocolversion"`
}

type BitcoinChain struct {
}

func NewBitcoinChain() *BitcoinChain {
	return &BitcoinChain{}
}

func (c BitcoinChain) GetClientVersion(ctx context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsoff.NewRequestMessage(1, "getnetworkinfo", nil)
	var info bitcoinNetworkInfo
	err := ep.UnwrapCallRPC(ctx, reqmsg, &info)
	if err != nil {
		return "", err
	}
	v := fmt.Sprintf("%d %s %d", info.Version, info.SubVersion, info.ProtocolVersion)
	return v, nil
}

func (c BitcoinChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

// update txid cache from mempool
// func (c BitcoinChain) updateMempoolPresenceCache(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) {
// 	redisClient, ok := m.RedisClient(presenceCacheRedisSelector(ep.Chain))
// 	if !ok {
// 		return
// 	}
// 	reqmsg := jsoff.NewRequestMessage(
// 		1, "getrawmempool", nil)

// 	var txids []string
// 	err := ep.UnwrapCallRPC(ctx, reqmsg, &txids)
// 	if err != nil {
// 		ep.Log().Warnf("getrawmempool error, %s", err)
// 		return
// 	}
// 	presenceCacheUpdate(
// 		ctx, redisClient,
// 		ep.Chain,
// 		txids,
// 		ep.Name,
// 		time.Second*600) // expire after 10 mins
// }

// func (c BitcoinChain) updateBlockPresenceCache(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint, blockHash string) {
// 	client, ok := m.RedisClient(presenceCacheRedisSelector(ep.Chain))
// 	if !ok {
// 		return
// 	}
// 	reqmsg := jsoff.NewRequestMessage(
// 		1, "getblock", []interface{}{blockHash})

// 	var blk bitcoinBlock
// 	err := ep.UnwrapCallRPC(ctx, reqmsg, &blk)
// 	if err != nil {
// 		ep.Log().Warnf("get block error, blockhash %s, %s", blockHash, err)
// 		return
// 	}
// 	presenceCacheUpdate(
// 		ctx, client,
// 		ep.Chain,
// 		blk.Tx,
// 		ep.Name,
// 		time.Second*1800) // expire after 30 mins
// }

func (c *BitcoinChain) GetBlockhead(ctx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "getblockchaininfo", nil)

	var chainInfo bitcoinBlockchainInfo
	err := ep.UnwrapCallRPC(ctx, reqmsg, &chainInfo)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: chainInfo.Blocks,
		Hash:   chainInfo.BestBlockhash,
	}
	return block, nil
}

func (c *BitcoinChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	//useCache := reqmsg.Method == "gettransaction" || reqmsg.Method == "getrawtransaction" || reqmsg.Method == "decoderawtransaction"
	useCache, resmsgFromCache := jsonrpcCacheFetchForMethods(
		ctx, m, chain, reqmsg,
		"gettransaction",
		"getrawtransaction",
		"decoderawtransaction")

	if resmsgFromCache != nil {
		return resmsgFromCache, nil
	}

	if ep, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg,
		"gettransaction",
		"getrawtransaction"); ok {
		retmsg, err := ep.CallRPC(ctx, reqmsg)
		if err == nil && useCache && retmsg.IsResult() {
			jsonrpcCacheUpdate(ctx, m, chain, reqmsg, retmsg.(*jsoff.ResultMessage), time.Minute*10)
		}
		return retmsg, err
	}

	if reqmsg.Method == "getblockhash" {
		if h, ok := c.findBlockHeight(reqmsg); ok {
			return m.DefaultRelayRPC(ctx, chain, reqmsg, h)
		}
	} else if reqmsg.Method == "getchaintips" || reqmsg.Method == "getblockchaininfo" {
		// select latest chaintips
		return m.DefaultRelayRPC(ctx, chain, reqmsg, 0)
	}

	retmsg, err := m.DefaultRelayRPC(ctx, chain, reqmsg, -1)
	if err == nil && useCache && retmsg.IsResult() {
		jsonrpcCacheUpdate(ctx, m, chain, reqmsg, retmsg.(*jsoff.ResultMessage), time.Minute*10)
	}
	return retmsg, err
}

func (c *BitcoinChain) findBlockHeight(reqmsg *jsoff.RequestMessage) (int, bool) {
	// the first argument is a integer number
	var bh struct {
		Height int
	}
	if err := jsoff.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height > 0 {
		return bh.Height, true
	}
	return 0, false
}
