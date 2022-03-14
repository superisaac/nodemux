package chains

// docsite: https://ethereum.org/en/developers/docs/apis/json-rpc/

import (
	"context"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"github.com/superisaac/nodemux/core"
	"reflect"
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

type ethereumHeadSub struct {
	Subscription string
	Result       ethereumBlock
}

type ethereumSubkey struct {
	EpName string
	Token  string
}

type EthereumChain struct {
	subTokens map[ethereumSubkey]bool
}

func NewEthereumChain() *EthereumChain {
	return &EthereumChain{
		subTokens: make(map[ethereumSubkey]bool),
	}
}

func (self EthereumChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	reqmsg := jsonz.NewRequestMessage(
		1, "web3_clientVersion", nil)
	var v string
	err := ep.UnwrapCallRPC(context, reqmsg, &v)
	if err != nil {
		ep.Log().Warnf("error call rpc web3_clientVersion %s", err)
		return "", err
	}
	return v, nil
}

func (self EthereumChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	if !ep.IsWebsocket() {
		return true, nil
	}

	// subscribe chaintip from is websocket
	go self.subscribeChaintip(context, m, ep)
	return false, nil
}

func (self *EthereumChain) GetChaintip(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	reqmsg := jsonz.NewRequestMessage(
		jsonz.NewUuid(), "eth_getBlockByNumber",
		[]interface{}{"latest", false})

	var bt ethereumBlock
	err := ep.UnwrapCallRPC(context, reqmsg, &bt)
	if err != nil {
		return nil, err
	}

	block := &nodemuxcore.Block{
		Height: bt.Height(),
		Hash:   bt.Hash,
	}

	if ep.Chaintip == nil || ep.Chaintip.Height != bt.Height() {
		if c, ok := m.RedisClient(presenceCacheRedisKey(ep.Chain)); ok {
			go presenceCacheUpdate(
				context, c,
				ep.Chain,
				bt.Transactions, ep.Name,
				time.Second*600) // expire after 10 mins
		}
	}
	return block, nil
}

func (self *EthereumChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
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
	return m.DefaultRelayRPC(ctx, chain, reqmsg, -5)
}

func (self *EthereumChain) findBlockHeight(reqmsg *jsonz.RequestMessage) (int, bool) {
	// the first argument is a hexlified block number or latest or pending
	var bh struct {
		Height string
	}
	if err := jsonz.DecodeParams(reqmsg.Params, &bh); err == nil && bh.Height != "" {
		if bh.Height == "latest" || bh.Height == "pending" {
			return 0, true
		}
		if height, err := hexutil.DecodeUint64(bh.Height); err == nil {
			return int(height), true
		}
	}
	return 0, false
}

func (self *EthereumChain) subscribeChaintip(rootCtx context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) {
	wsClient, ok := ep.RPCClient().(*jsonzhttp.WSClient)

	if !ok {
		ep.Log().Panicf("client is not websocket, client is %s", reflect.TypeOf(ep.RPCClient()))
		return
		//return errors.New("client is not websocket")
	}

	wsClient.OnMessage(func(msg jsonz.Message) {
		ntf, ok := msg.(*jsonz.NotifyMessage)
		if !ok && ntf.Method != "eth_subscription" || len(ntf.Params) == 0 {
			return
		}
		var headSub ethereumHeadSub
		err := jsonz.DecodeInterface(ntf.Params[0], &headSub)
		if err != nil {
			ep.Log().Warnf("decode head sub error %s", err)
		} else {
			// match Subscription against sub token
			subkey := ethereumSubkey{EpName: ep.Name, Token: headSub.Subscription}
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
				Chaintip:     headBlock,
				Unhealthy:    false,
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
				Unhealthy:    true,
			}
			m.Chainhub().Pub() <- bs
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (self *EthereumChain) connectAndSub(rootCtx context.Context, wsClient *jsonzhttp.WSClient, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// connect websocket
	err := wsClient.Connect(ctx)
	if err != nil {
		return err
	}

	// request chaintip
	headBlock, err := self.GetChaintip(ctx, m, ep)
	if err != nil {
		return err
	}
	if headBlock != nil {
		bs := nodemuxcore.ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Chaintip:     headBlock,
			Unhealthy:    false,
		}
		m.Chainhub().Pub() <- bs
	}

	// send sub command
	var subscribeToken string
	submsg := jsonz.NewRequestMessage(
		jsonz.NewUuid(), "eth_subscribe",
		[]interface{}{"newHeads"})

	err = ep.UnwrapCallRPC(ctx, submsg, &subscribeToken)
	if err != nil {
		return err
	}
	subkey := ethereumSubkey{
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
