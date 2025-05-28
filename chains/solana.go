package chains

import (
	"context"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/nodemux/core"
	"net/http"
	"time"
)

var (
	solanaCachableMethods map[string]time.Duration = map[string]time.Duration{
		"getBlock":       time.Second * 5,
		"getSlot":        time.Second * 4,
		"getTransaction": time.Second * 600,
	}
)

type SolanaChain struct {
}

func NewSolanaChain() *SolanaChain {
	return &SolanaChain{}
}

func (c SolanaChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (c SolanaChain) StartSync(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (bool, error) {
	return true, nil
}

func (c *SolanaChain) GetBlockhead(context context.Context, m *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	config := map[string]string{"commitement": "confirmed"}
	reqmsg := jsoff.NewRequestMessage(
		1, "getSlot", []interface{}{config})

	var slot int
	err := ep.UnwrapCallRPC(context, reqmsg, &slot)
	if err != nil {
		return nil, err
	}
	block := &nodemuxcore.Block{
		Height: slot,
	}
	return block, nil
}

func (c *SolanaChain) DelegateRPC(ctx context.Context, m *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, reqmsg *jsoff.RequestMessage, r *http.Request) (jsoff.Message, error) {
	useCache := false
	cacheExpire := time.Second * 60
	if exp, ok := solanaCachableMethods[reqmsg.Method]; ok {
		useCache = true
		cacheExpire = exp
		if resmsgFromCache, found := jsonrpcCacheFetch(ctx, m, chain, reqmsg); found {
			reqmsg.Log().Infof("get result from cache")
			return resmsgFromCache, nil
		}
	}

	retmsg, ep, err := m.DefaultRelayRPCTakingEndpoint(ctx, chain, reqmsg, -10)
	if err == nil && ep != nil && useCache && retmsg.IsResult() {
		jsonrpcCacheUpdate(ctx, m, ep, chain, reqmsg, retmsg.(*jsoff.ResultMessage), cacheExpire)
	}
	return retmsg, err

	// return m.DefaultRelayRPC(ctx, chain, reqmsg, -10)
}
