package chains

// docsite: https://ethereum.org/en/developers/docs/apis/json-rpc/

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"reflect"
	"time"
)

var (
	// refer to https://eth.wiki/json-rpc/API
	allowedMethods map[string]bool = map[string]bool{
		"web3_clientVersion": true,
		"web3_sha3":          true,

		"net_version":   true,
		"net_peerCount": true,
		"net_listening": true,

		"eth_gasPrice":        true,
		"eth_estimateGas":     true,
		"eth_protocolVersion": true,
		"eth_syncing":         true,
		"eth_coinbase":        true,

		"eth_blockNumber":                         true,
		"eth_getBlockByNumber":                    true,
		"eth_getBlockByHash":                      true,
		"eth_subscribe":                           true,
		"eth_unsubscribe":                         true,
		"eth_getTransactionByHash":                true,
		"eth_getTransactionCount":                 true,
		"eth_getTransactionByBlockHashAndIndex":   true,
		"eth_getTransactionByBlockNumberAndIndex": true,
		"eth_getTransactionReceipt":               true,
		"eth_getUncleByBlockHashAndIndex":         true,
		"eth_getUncleByBlockNumberAndIndex":       true,

		"eth_getBalance":         true,
		"eth_getStorageAt":       true,
		"eth_call":               true,
		"eth_getCode":            true,
		"eth_sendRawTransaction": true,
		"eth_getLogs":            true,

		// wallet/account related RPCs are not supported by default

		"eth_getCompilers":    true,
		"eth_compileLLL":      true,
		"eth_compileSolidity": true,
		"eth_compileSerpent":  true,

		// mining
		"eth_mining":         true,
		"eth_hashrate":       true,
		"eth_getWork":        true,
		"eth_submitWork":     true,
		"eth_submitHashrate": true,

		// filter related RPCs are not supported by default

		"parity_getBlockReceipts": true,

		"debug_traceTransaction": true,

		// trace
		"trace_call":                    true,
		"trace_callMany":                true,
		"trace_rawTransaction":          true,
		"trace_replayBlockTransactions": true,
		"trace_replayTransaction":       true,
		"trace_block":                   true,
		"trace_get":                     true,
		"trace_transaction":             true,
	}
)

type web3Block struct {
	Number       string
	Hash         string
	Transactions []string `json:"transactions"`

	// private fields
	height int
}

func (self *web3Block) Height() int {
	if self.height <= 0 {
		height, err := hexutil.DecodeUint64(self.Number)
		if err != nil {
			panic(err)
		}
		self.height = int(height)
	}
	return self.height
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

func (self Web3Chain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jlib.NewRequestMessage(
		1, "web3_clientVersion", nil)
	var v string
	err := ep.UnwrapCallRPC(context, reqmsg, &v)
	if err != nil {
		ep.Log().Warnf("error call rpc web3_clientVersion %s", err)
		return "", err
	}
	return v, nil
}

func (self Web3Chain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	if !ep.IsWebsocket() {
		return true, nil
	}

	// subscribe chaintip from is websocket
	go self.subscribeBlockhead(context, m, ep)
	return false, nil
}

func (self *Web3Chain) GetBlockhead(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jlib.NewRequestMessage(
		jlib.NewUuid(), "eth_getBlockByNumber",
		[]interface{}{"latest", false})

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

func (self *Web3Chain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jlib.RequestMessage, r *http.Request) (jlib.Message, error) {
	if allowed, ok := allowedMethods[reqmsg.Method]; !ok || !allowed {
		reqmsg.Log().Warnf("relayer method not supported %s", reqmsg.Method)
		return jlib.ErrMethodNotFound.ToMessage(reqmsg), nil
	}

	if reqmsg.Method == "web3_clientVersion" {
		return jlib.NewResultMessage(reqmsg, "Web3/1.0.0"), nil
	}

	//useCache := reqmsg.Method == "eth_getTransactionByHash" || reqmsg.Method == "eth_getTransactionReceipt"
	useCache, resmsgFromCache := jsonrpcCacheFetchForMethods(
		ctx, m, chain, reqmsg,
		"eth_getTransactionByHash",
		"eth_getTransactionReceipt")

	if resmsgFromCache != nil {
		return resmsgFromCache, nil
	}

	if endpoint, ok := presenceCacheMatchRequest(
		ctx, m, chain, reqmsg,
		"eth_getTransactionByHash",
		"eth_getTransactionReceipt"); ok {
		return endpoint.CallRPC(ctx, reqmsg)
	}

	if reqmsg.Method == "eth_getBlockByNumber" {
		if h, ok := self.findBlockHeight(reqmsg); ok {
			return m.DefaultRelayRPC(ctx, chain, reqmsg, h)
		}
	}
	//return m.DefaultRelayRPC(ctx, chain, reqmsg, -2)
	retmsg, err := m.DefaultRelayRPC(ctx, chain, reqmsg, -2)
	if err == nil && useCache && retmsg.IsResult() {
		jsonrpcCacheUpdate(ctx, m, chain, reqmsg, retmsg.(*jlib.ResultMessage), time.Second*600)
	}
	return retmsg, nil
}

func (self *Web3Chain) findBlockHeight(reqmsg *jlib.RequestMessage) (int, bool) {
	// the first argument is a hexlified block number or latest or pending
	var bh struct {
		Height string
	}
	if err := jlib.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height != "" {
		if bh.Height == "latest" || bh.Height == "pending" {
			return 0, true
		}
		if height, err := hexutil.DecodeUint64(bh.Height); err == nil {
			return int(height), true
		}
	}
	return 0, false
}

func (self *Web3Chain) subscribeBlockhead(rootCtx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) {
	wsClient, ok := ep.JSONRPCRelayer().(*jlibhttp.WSClient)

	if !ok {
		ep.Log().Panicf("client is not websocket, client is %s", reflect.TypeOf(ep.JSONRPCRelayer()))
		return
		//return errors.New("client is not websocket")
	}

	wsClient.OnMessage(func(msg jlib.Message) {
		ntf, ok := msg.(*jlib.NotifyMessage)
		if !ok && ntf.Method != "eth_subscription" || len(ntf.Params) == 0 {
			return
		}
		var headSub web3HeadSub
		err := jlib.DecodeInterface(ntf.Params[0], &headSub)
		if err != nil {
			ep.Log().Warnf("decode head sub error %s", err)
		} else {
			// match Subscription against sub token
			subkey := web3Subkey{EpName: ep.Name, Token: headSub.Subscription}
			if _, ok := self.subTokens[subkey]; !ok {
				ep.Log().Warnf("subscription %s not found amount %#v",
					headSub.Subscription,
					self.subTokens)
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
		err := self.connectAndSub(rootCtx, wsClient, m, ep)
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

func (self *Web3Chain) connectAndSub(rootCtx context.Context, wsClient *jlibhttp.WSClient, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// connect websocket
	err := wsClient.Connect(ctx)
	if err != nil {
		return err
	}

	// request chaintip
	headBlock, err := self.GetBlockhead(ctx, m, ep)
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
	submsg := jlib.NewRequestMessage(
		jlib.NewUuid(), "eth_subscribe",
		[]interface{}{"newHeads"})

	err = ep.UnwrapCallRPC(ctx, submsg, &subscribeToken)
	if err != nil {
		return err
	}
	subkey := web3Subkey{
		EpName: ep.Name,
		Token:  subscribeToken,
	}

	self.subTokens[subkey] = true
	ep.Log().Infof("eth got subscrib token %s", subscribeToken)
	defer func() {
		delete(self.subTokens, subkey)
	}()

	return wsClient.Wait()

}
