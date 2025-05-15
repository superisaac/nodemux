package nodemuxcore

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff"
	"net/http"
	"sync"
)

// singleton vars and methods
var (
	_instance *Multiplexer
)

var (
	ErrNotAvailable = &jsoff.RPCError{Code: -32060, Message: "not available"}
)

func GetMultiplexer() *Multiplexer {
	if _instance == nil {
		log.Panicf("Multiplexer instance is nil")
	}
	return _instance
}

func SetMultiplexer(b *Multiplexer) {
	if _instance != nil {
		_instance.StopSync()
	}
	_instance = b
}

func NewMultiplexer() *Multiplexer {
	m := new(Multiplexer)
	m.chainHub = NewMemoryChainhub()
	m.Reset()
	return m
}

func MultiplexerFromConfig(nbcfg *NodemuxConfig) *Multiplexer {
	m := NewMultiplexer()
	m.LoadFromConfig(nbcfg)

	if rdb, ok := m.RedisClient("default"); ok {
		// sync source must be a redis URL
		log.Infof("using redis stream store")
		chainHub, err := NewRedisStreamChainhub(rdb)
		//chainHub, err := NewRedisChainhub(rdb)
		if err != nil {
			panic(err)
		}
		m.chainHub = chainHub
	} else {
		log.Info("using memory store")
	}
	return m
}

func (m *Multiplexer) Reset() {
	m.nameIndex = make(map[string]*Endpoint)
	m.chainIndex = make(map[ChainRef]*EndpointSet)
	m.redisClients = make(map[string]*redis.Client)
}

func (m Multiplexer) Get(epName string) (*Endpoint, bool) {
	ep, ok := m.nameIndex[epName]
	return ep, ok
}

func (m Multiplexer) MustGet(epName string) *Endpoint {
	if ep, ok := m.Get(epName); ok {
		return ep
	}
	log.Panicf("fail to get endpoint %s", epName)
	return nil
}

func (m Multiplexer) Chainhub() Chainhub {
	return m.chainHub
}

func (m *Multiplexer) Add(endpoint *Endpoint) bool {
	if _, exist := m.nameIndex[endpoint.Name]; exist {
		// already exist
		log.Warnf("endpoint %s already exist", endpoint.Name)
		return false
	}
	m.nameIndex[endpoint.Name] = endpoint

	if eps, ok := m.chainIndex[endpoint.Chain]; ok {
		eps.Add(endpoint)
	} else {
		eps := NewEndpointSet()
		m.chainIndex[endpoint.Chain] = eps
		eps.Add(endpoint)
	}
	return true
}

func (m *Multiplexer) Select(chain ChainRef, method string) (*Endpoint, bool) {
	if eps, ok := m.chainIndex[chain]; ok {
		if epName, ok := eps.WeightedRandom(); ok {
			ep := eps.MustGet(epName)
			if ep.Available(method, 0) {
				return ep, true
			}

			// select endpoint sequentially
			for _, ep := range eps.items {
				if ep.Available(method, 0) {
					return ep, true
				}
			}
		}
	}
	return nil, false
}

func (m *Multiplexer) AllHealthyEndpoints(chain ChainRef, method string, height int) []*Endpoint {
	if endpoints, ok := m.chainIndex[chain]; ok {
		healthyEndpoints := make([]*Endpoint, 0)
		for _, ep := range endpoints.items {
			if ep.Available(method, height) {
				healthyEndpoints = append(healthyEndpoints, ep)
			}
		}
		return healthyEndpoints
	}
	return nil
}

func (m *Multiplexer) SelectEndpointByName(chain ChainRef, name string, method string) *Endpoint {
	if endpoints, ok := m.chainIndex[chain]; ok {
		for _, ep := range endpoints.items {
			if ep.Available(method, 0) && ep.Name == name {
				return ep
			}
		}
	}
	return nil
}

func (m *Multiplexer) SelectEndpointFromHttp(chain ChainRef, method string, r *http.Request) *Endpoint {
	selectNode := r.Header.Get("X-Nodemux-Select")
	if selectNode == "" {
		return nil
	}
	return m.SelectEndpointByName(chain, selectNode, method)
}

func (m *Multiplexer) SelectOverHeight(chain ChainRef, method string, heightSpec int) (*Endpoint, bool) {
	if endpoints, ok := m.chainIndex[chain]; ok {
		height := heightSpec
		if heightSpec <= 0 {
			height = endpoints.maxTipHeight + heightSpec
		}

		// select a random endpoint by weights, if it's not available then select by sequence
		if epName, ok := endpoints.WeightedRandom(); ok {
			ep := endpoints.MustGet(epName)
			if ep.Available(method, height) {
				return ep, true
			}

			for _, ep := range endpoints.items {
				if ep.Available(method, height) {
					return ep, true
				}
			}
		}
	}
	return nil, false
}

func (m *Multiplexer) RequestCacheKeys(chain ChainRef, reqmsg *jsoff.RequestMessage, prefix string, heightSpec int) []string {
	if endpoints, ok := m.chainIndex[chain]; ok {
		height := heightSpec
		if heightSpec <= 0 {
			height = endpoints.maxTipHeight + heightSpec
		}

		var cacheKeys []string
		for _, ep := range endpoints.items {
			if ep.Available(reqmsg.Method, height) {
				cacheKeys = append(cacheKeys, reqmsg.CacheKey(fmt.Sprintf("%s%s/", prefix, ep.Name)))
			}
		}
		return cacheKeys
	}
	return nil
}

func (m *Multiplexer) SelectWebsocketEndpoint(chain ChainRef, method string, heightSpec int) (ep1 *Endpoint, found bool) {
	if endpoints, ok := m.chainIndex[chain]; ok {
		height := heightSpec
		if heightSpec < 0 {
			height = endpoints.maxTipHeight + heightSpec
		}

		if epName, ok := endpoints.WeightedRandom(); ok {
			ep := endpoints.MustGet(epName)
			if ep.HasWebsocket() && ep.Available(method, height) {
				return ep, true
			}
			for _, ep := range endpoints.items {
				if ep.HasWebsocket() && ep.Available(method, height) {
					return ep, true
				}
			}
		}
	}
	return nil, false
}

func (m *Multiplexer) LoadFromConfig(nbcfg *NodemuxConfig) {
	m.cfg = nbcfg
	for name, epcfg := range nbcfg.Endpoints {
		chainref, err := ParseChain(epcfg.Chain)
		if err != nil {
			panic(err)
		}
		if support, _ := GetDelegatorFactory().SupportChain(chainref.Namespace); !support {
			panic(fmt.Sprintf("chain %s not supported", chainref))
		}

		ep := NewEndpoint(name, epcfg)
		m.Add(ep)
	}
}

func (m *Multiplexer) BroadcastRPC(
	rootCtx context.Context,
	chain ChainRef,
	reqmsg *jsoff.RequestMessage,
	overHeight int) []RPCResult {
	eps := m.AllHealthyEndpoints(chain, reqmsg.Method, overHeight)
	if len(eps) == 0 {
		return nil
	}

	wg := new(sync.WaitGroup)
	results := make([]RPCResult, 0)
	lock := sync.RWMutex{}

	for _, ep := range eps {
		wg.Add(1)
		go func(ep *Endpoint) {
			resmsg, err := ep.CallRPC(rootCtx, reqmsg)
			lock.Lock()
			defer lock.Unlock()
			results = append(results, RPCResult{
				Response: resmsg,
				Endpoint: ep,
				Err:      err,
			})
			wg.Done()
		}(ep)
	}
	wg.Wait()
	return results
}

func (m *Multiplexer) CallEndpointRPC(rootCtx context.Context, ep *Endpoint, reqmsg *jsoff.RequestMessage) (jsoff.Message, error) {
	resmsg, err := ep.CallRPC(rootCtx, reqmsg)
	if err != nil {
		return resmsg, err
	}
	if responseMsg, ok := resmsg.(jsoff.ResponseMessage); ok {
		responseMsg.ResponseHeader().Set("X-Real-Endpoint", ep.Name)
	}
	return resmsg, err
}

func (m *Multiplexer) DefaultRelayRPC(
	rootCtx context.Context,
	chain ChainRef,
	reqmsg *jsoff.RequestMessage,
	overHeight int) (jsoff.Message, error) {
	ep, found := m.SelectOverHeight(chain, reqmsg.Method, overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return m.DefaultRelayRPC(rootCtx, chain, reqmsg, -2)
		}
		return ErrNotAvailable.ToMessage(reqmsg), nil
	}
	return m.CallEndpointRPC(rootCtx, ep, reqmsg)
}

func (m *Multiplexer) DefaultRelayRPCTakingEndpoint(
	rootCtx context.Context,
	chain ChainRef,
	reqmsg *jsoff.RequestMessage,
	overHeight int) (jsoff.Message, *Endpoint, error) {
	ep, found := m.SelectOverHeight(chain, reqmsg.Method, overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return m.DefaultRelayRPCTakingEndpoint(rootCtx, chain, reqmsg, -2)
		}
		return ErrNotAvailable.ToMessage(reqmsg), nil, nil
	}
	msg, err := m.CallEndpointRPC(rootCtx, ep, reqmsg)
	return msg, ep, err
}

// Pipe the request to response
func (m *Multiplexer) DefaultPipeREST(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := m.SelectOverHeight(chain, path, overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return m.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
		}
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}

func (m *Multiplexer) DefaultPipeGraphQL(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := m.SelectOverHeight(chain, "", overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return m.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -2)
		}
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}

func (m Multiplexer) ListEndpointInfos() []EndpointInfo {
	infos := make([]EndpointInfo, 0)
	for _, ep := range m.nameIndex {
		infos = append(infos, ep.Info())
	}
	return infos
}
