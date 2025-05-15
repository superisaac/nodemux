package chains

// docsite: https://ethereum.org/en/developers/docs/apis/json-rpc/

import (
	"context"
	"net/http"
	"strings"
	"time"
	// "fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/jsoff/net"
	"github.com/superisaac/nodemux/core"
)

var (
	// // refer to https://eth.wiki/json-rpc/API
	// allowedMethods map[string]bool = map[string]bool{
	// 	"web3_clientVersion": true,
	// 	"web3_sha3":          true,

	// 	"net_version":   true,
	// 	"net_peerCount": true,
	// 	"net_listening": true,

	// 	"eth_gasPrice":        true,
	// 	"eth_estimateGas":     true,
	// 	"eth_protocolVersion": true,
	// 	"eth_syncing":         true,
	// 	"eth_coinbase":        true,
	// 	"eth_chainId":         true,

	// 	"eth_blockNumber":                         true,
	// 	"eth_getBlockByNumber":                    true,
	// 	"eth_getBlockByHash":                      true,
	// 	"eth_subscribe":                           true,
	// 	"eth_unsubscribe":                         true,
	// 	"eth_getTransactionByHash":                true,
	// 	"eth_getTransactionCount":                 true,
	// 	"eth_getTransactionByBlockHashAndIndex":   true,
	// 	"eth_getTransactionByBlockNumberAndIndex": true,
	// 	"eth_getTransactionReceipt":               true,
	// 	"eth_getUncleByBlockHashAndIndex":         true,
	// 	"eth_getUncleByBlockNumberAndIndex":       true,

	// 	"eth_getBalance":         true,
	// 	"eth_getStorageAt":       true,
	// 	"eth_call":               true,
	// 	"eth_getCode":            true,
	// 	"eth_sendRawTransaction": true,
	// 	"eth_getLogs":            true,

	// 	// wallet/account related RPCs are not supported by default

	// 	"eth_getCompilers":    true,
	// 	"eth_compileLLL":      true,
	// 	"eth_compileSolidity": true,
	// 	"eth_compileSerpent":  true,

	// 	// mining
	// 	"eth_mining":         true,
	// 	"eth_hashrate":       true,
	// 	"eth_getWork":        true,
	// 	"eth_submitWork":     true,
	// 	"eth_submitHashrate": true,

	// 	// filter related RPCs are not supported by default

	// 	"parity_getBlockReceipts": true,

	// 	"debug_traceTransaction": true,

	// 	// trace
	// 	"trace_call":                    true,
	// 	"trace_callMany":                true,
	// 	"trace_rawTransaction":          true,
	// 	"trace_replayBlockTransactions": true,
	// 	"trace_replayTransaction":       true,
	// 	"trace_block":                   true,
	// 	"trace_get":                     true,
	// 	"trace_transaction":             true,
	// }

	web3CachableMethods map[string]time.Duration = map[string]time.Duration{
		"eth_getBlockByNumber":                    time.Second * 60,
		"eth_getBlockByHash":                      time.Second * 600,
		"eth_getTransactionByHash":                time.Second * 600,
		"eth_getTransactionCount":                 time.Second * 5,
		"eth_getTransactionByBlockHashAndIndex":   time.Second * 30,
		"eth_getTransactionByBlockNumberAndIndex": time.Second * 30,
		"eth_getTransactionReceipt":               time.Second * 10,
	}
)

type web3Block struct {
	Number       string
	Hash         string
	Transactions []string `json:"transactions"`

	// private fields
	height int
}

func (blk *web3Block) Height() int {
	if blk.height <= 0 {
		height, err := hexutil.DecodeUint64(blk.Number)
		if err != nil {
			panic(err)
		}
		blk.height = int(height)
	}
	return blk.height
}

type web3HeadSub struct {
	Subscription string
	Result       web3Block
}

type web3Subkey struct {
	EpName string
	Token  string
}

type Web3Chain struct {
	subTokens map[web3Subkey]bool
}

func NewWeb3Chain() *Web3Chain {
	return &Web3Chain{
		subTokens: make(map[web3Subkey]bool),
	}
}

func (c Web3Chain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsoff.NewRequestMessage(
		1, "web3_clientVersion", nil)
	var v string
	err := ep.UnwrapCallRPC(context, reqmsg, &v)
	if err != nil {
		ep.Log().Warnf("error call rpc web3_clientVersion %s", err)
		return "", err
	}
	return v, nil
}

func (c Web3Chain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	if !ep.HasWebsocket() {
		return true, nil
	}

	// subscribe chaintip from is websocket
	go c.subscribeBlockhead(context, m, ep)
	return false, nil
}

func (c *Web3Chain) GetBlockhead(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsoff.NewRequestMessage(
		jsoff.NewUuid(), "eth_getBlockByNumber",
		[]any{"latest", false})

	var bt web3Block
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Height(),
		Hash:   bt.Hash,
	}

	if ep.Blockhead == nil || ep.Blockhead.Height != bt.Height() {
		if c, ok := m.RedisClient(presenceCacheRedisSelector(ep.Chain)); ok {
			go presenceCacheUpdate(
				context, c,
				ep.Chain,
				bt.Transactions, ep.Name,
				time.Second*600) // expire after 10 mins
		}
	}
	return block, nil
}

func (c *Web3Chain) sendRawTransaction(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage) (jsoff.Message, error) {
	resMsgs := m.BroadcastRPC(ctx, chain, reqmsg, -10)
	if len(resMsgs) == 0 {
		return m.DefaultRelayRPC(ctx, chain, reqmsg, -5)
	}

	// try find the correct response
	// return the first correct response
	for _, res := range resMsgs {
		if res.Err == nil && res.Response.IsResult() {
			return res.Response, nil
		}
	}

	// return the first -32000, already known result
	for _, res := range resMsgs {
		if res.Err == nil && res.Response.IsError() && res.Response.MustError().Code == -32000 {
			return res.Response, nil
		}
	}

	// return the first error msg
	for _, res := range resMsgs {
		if res.Err == nil && res.Response.IsError() {
			return res.Response, nil
		}
	}

	// return the first item that has a message
	for _, res := range resMsgs {
		if res.Err == nil && res.Response != nil {
			return res.Response, nil
		}
	}
	// reutrn error
	return nil, resMsgs[0].Err
}

func (c *Web3Chain) getBlockByNumber(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	useCache := true
	cacheExpire := time.Second * 30
	heightSpec := -2
	if h, ok := c.findBlockHeight(reqmsg); ok {
		heightSpec = h
	}

	if heightSpec <= 0 {
		cacheExpire = time.Second * 4
	}

	retmsg, ep, err := m.DefaultRelayRPCTakingEndpoint(ctx, chain, reqmsg, heightSpec)
	//fmt.Printf("ret %#v, %#v, %#v\n", retmsg, ep, err)
	if err == nil && ep != nil && useCache && retmsg.IsResult() {
		jsonrpcCacheUpdate(ctx, m, ep, chain, reqmsg, retmsg.(*jsoff.ResultMessage), cacheExpire)
	}
	return retmsg, err
}

func (c *Web3Chain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	// if allowed, ok := allowedMethods[reqmsg.Method]; !ok || !allowed {
	// 	reqmsg.Log().Warnf("relayer method not supported %s", reqmsg.Method)
	// 	return jsoff.ErrMethodNotFound.ToMessage(reqmsg), nil
	// }

	if reqmsg.Method == "web3_clientVersion" {
		return jsoff.NewResultMessage(reqmsg, "Web3/1.0.0"), nil
	}

	useCache := false
	cacheExpire := time.Second * 60
	if exp, ok := web3CachableMethods[reqmsg.Method]; ok {
		useCache = true
		cacheExpire = exp
		if resmsgFromCache, found := jsonrpcCacheFetch(ctx, m, chain, reqmsg); found {
			reqmsg.Log().Infof("get result from cache")
			return resmsgFromCache, nil
		}
	}

	if reqmsg.Method == "eth_getBlockByNumber" {
		return c.getBlockByNumber(ctx, m, chain, reqmsg, r)
	}

	if endpoint, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg,
		"eth_getTransactionByHash",
		"eth_getTransactionReceipt",
	); ok {
		retmsg, err := endpoint.CallRPC(ctx, reqmsg)
		if err == nil && useCache && retmsg.IsResult() {
			jsonrpcCacheUpdate(ctx, m, endpoint, chain, reqmsg, retmsg.(*jsoff.ResultMessage), cacheExpire)
		}
		return retmsg, nil
	}

	if reqmsg.Method == "eth_sendRawTransaction" {
		// broadcast raw transactions to all endpoints
		return c.sendRawTransaction(ctx, m, chain, reqmsg)
	}

	heightSpec := -2

	retmsg, ep, err := m.DefaultRelayRPCTakingEndpoint(ctx, chain, reqmsg, heightSpec)
	if err == nil && useCache && retmsg.IsResult() {
		jsonrpcCacheUpdate(ctx, m, ep, chain, reqmsg, retmsg.(*jsoff.ResultMessage), cacheExpire)
	}
	if err == nil && reqmsg.Method == "eth_getTransactionReceipt" {
		if respMsg, ok := retmsg.(*jsoff.ResultMessage); ok && respMsg.Result == nil {
			// viaEp := respMsg.ResponseHeader().Get("X-Real-Endpoint")
			respMsg.Log().Warnf("null transaction receipt %s", ep.Name)
		}
	}
	return retmsg, err
}

func (c *Web3Chain) findBlockHeight(reqmsg *jsoff.RequestMessage) (int, bool) {
	// the first argument is a hexlified block number or latest or pending
	var bh struct {
		Height string
	}
	if err := jsoff.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height != "" {
		if strings.HasPrefix(bh.Height, "0x") {
			if height, err := hexutil.DecodeUint64(bh.Height); err == nil {
				return int(height), true
			}
		} else {
			return 0, true
		}
		// if bh.Height == "latest" || bh.Height == "pending" {
		// 	return 0, true
		// }
	}
	return 0, false
}

func (c *Web3Chain) subscribeBlockhead(rootCtx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) {
	wsClient, ok := ep.NewJSONRPCWSClient()
	if !ok {
		ep.Log().Panicf("endpoint has no websocket client, %s", ep.Name)
		return
	}

	wsClient.OnMessage(func(msg jsoff.Message) {
		ntf, ok := msg.(*jsoff.NotifyMessage)
		if !ok || ntf == nil {
			return
		}
		if ntf.Method != "eth_subscription" || len(ntf.Params) == 0 {
			return
		}
		var headSub web3HeadSub
		err := jsoff.DecodeInterface(ntf.Params[0], &headSub)
		if err != nil {
			ep.Log().Warnf("decode head sub error %s", err)
		} else {
			// match Subscription against sub token
			subkey := web3Subkey{EpName: ep.Name, Token: headSub.Subscription}
			if _, ok := c.subTokens[subkey]; !ok {
				ep.Log().Warnf("subscription %s not found amount %#v",
					headSub.Subscription,
					c.subTokens)
				return
			}
			headBlock := &nodemuxcore.Block{
				Height: headSub.Result.Height(),
				Hash:   headSub.Result.Hash,
			}
			bs := nodemuxcore.ChainStatus{
				EndpointName: ep.Name,
				Chain:        ep.Chain,
				Healthy:      true,
				Blockhead:    headBlock,
			}
			m.Chainhub().Pub() <- bs
		}
	}) // end of wsClient.OnMessage

	for {
		err := c.connectAndSub(rootCtx, wsClient, m, ep)
		if err != nil {
			ep.Log().Warnf("connsub error %s, retrying", err)
			bs := nodemuxcore.ChainStatus{
				EndpointName: ep.Name,
				Chain:        ep.Chain,
				Healthy:      false,
			}
			m.Chainhub().Pub() <- bs
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Web3Chain) connectAndSub(rootCtx context.Context, wsClient *jsoffnet.WSClient, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) error {
	connectCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// connect websocket
	err := wsClient.Connect(connectCtx)
	if err != nil {
		return err
	}

	// request chaintip
	headBlock, err := c.GetBlockhead(connectCtx, m, ep)
	if err != nil {
		return err
	}
	if headBlock != nil {
		bs := nodemuxcore.ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Healthy:      true,
			Blockhead:    headBlock,
		}
		m.Chainhub().Pub() <- bs
	}

	// send sub command
	var subscribeToken string
	submsg := jsoff.NewRequestMessage(
		jsoff.NewUuid(), "eth_subscribe",
		[]any{"newHeads"})

	err = ep.UnwrapCallRPC(connectCtx, submsg, &subscribeToken)
	if err != nil {
		return err
	}
	subkey := web3Subkey{
		EpName: ep.Name,
		Token:  subscribeToken,
	}

	c.subTokens[subkey] = true
	ep.Log().Infof("eth got subscrib token %s", subscribeToken)
	defer func() {
		delete(c.subTokens, subkey)
	}()

	return wsClient.Wait()
}
