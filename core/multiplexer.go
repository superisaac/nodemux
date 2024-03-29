package nodemuxcore

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff"
	"net/http"
	//"sync"
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

func (self *Multiplexer) Reset() {
	self.nameIndex = make(map[string]*Endpoint)
	self.chainIndex = make(map[ChainRef]*EndpointSet)
	self.redisClients = make(map[string]*redis.Client)
}

func (self Multiplexer) Get(epName string) (*Endpoint, bool) {
	ep, ok := self.nameIndex[epName]
	return ep, ok
}

func (self Multiplexer) MustGet(epName string) *Endpoint {
	if ep, ok := self.Get(epName); ok {
		return ep
	}
	log.Panicf("fail to get endpoint %s", epName)
	return nil
}

func (self Multiplexer) Chainhub() Chainhub {
	return self.chainHub
}

func (self *Multiplexer) Add(endpoint *Endpoint) bool {
	if _, exist := self.nameIndex[endpoint.Name]; exist {
		// already exist
		log.Warnf("endpoint %s already exist", endpoint.Name)
		return false
	}
	self.nameIndex[endpoint.Name] = endpoint

	if eps, ok := self.chainIndex[endpoint.Chain]; ok {
		eps.Add(endpoint)
	} else {
		eps := NewEndpointSet()
		self.chainIndex[endpoint.Chain] = eps
		eps.Add(endpoint)
	}
	return true
}

func (self *Multiplexer) Select(chain ChainRef, method string) (*Endpoint, bool) {
	if eps, ok := self.chainIndex[chain]; ok {
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

func (self *Multiplexer) SelectOverHeight(chain ChainRef, method string, heightSpec int) (*Endpoint, bool) {
	if endpoints, ok := self.chainIndex[chain]; ok {
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

func (self *Multiplexer) SelectWebsocketEndpoint(chain ChainRef, method string, heightSpec int) (ep1 *Endpoint, found bool) {
	if endpoints, ok := self.chainIndex[chain]; ok {
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

func (self *Multiplexer) LoadFromConfig(nbcfg *NodemuxConfig) {
	self.cfg = nbcfg
	for name, epcfg := range nbcfg.Endpoints {
		chainref, err := ParseChain(epcfg.Chain)
		if err != nil {
			panic(err)
		}
		if support, _ := GetDelegatorFactory().SupportChain(chainref.Namespace); !support {
			panic(fmt.Sprintf("chain %s not supported", chainref))
		}

		ep := NewEndpoint(name, epcfg)
		self.Add(ep)
	}
}

func (self *Multiplexer) DefaultRelayRPC(
	rootCtx context.Context,
	chain ChainRef,
	reqmsg *jsoff.RequestMessage,
	overHeight int) (jsoff.Message, error) {
	ep, found := self.SelectOverHeight(chain, reqmsg.Method, overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return self.DefaultRelayRPC(rootCtx, chain, reqmsg, -2)
		}
		return ErrNotAvailable.ToMessage(reqmsg), nil
	}
	resmsg, err := ep.CallRPC(rootCtx, reqmsg)
	if err != nil {
		return resmsg, err
	}
	if responseMsg, ok := resmsg.(jsoff.ResponseMessage); ok {
		responseMsg.ResponseHeader().Set("X-Real-Endpoint", ep.Name)
	}
	return resmsg, err
}

// Pipe the request to response
func (self *Multiplexer) DefaultPipeREST(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := self.SelectOverHeight(chain, path, overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return self.DefaultPipeREST(rootCtx, chain, path, w, r, -2)
		}
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}

func (self *Multiplexer) DefaultPipeGraphQL(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := self.SelectOverHeight(chain, "", overHeight)
	if !found {
		if overHeight > 0 {
			// if not find then relay to any healthy endpoint
			return self.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -2)
		}
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}

func (self Multiplexer) ListEndpointInfos() []EndpointInfo {
	infos := make([]EndpointInfo, 0)
	for _, ep := range self.nameIndex {
		infos = append(infos, ep.Info())
	}
	return infos
}
