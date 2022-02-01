package nodemuxcore

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonz"
	"net/http"
	//"sync"
)

var (
	_instance *Multiplexer
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
	b := new(Multiplexer)
	b.chainHub = NewLocalChainhub()
	b.Reset()
	return b
}

func (self *Multiplexer) Reset() {
	self.nameIndex = make(map[string]*Endpoint)
	self.chainIndex = make(map[ChainRef]*EndpointSet)
}

func (self Multiplexer) Get(epName string) (*Endpoint, bool) {
	ep, ok := self.nameIndex[epName]
	return ep, ok
}

func (self *Multiplexer) Add(endpoint *Endpoint) bool {
	if _, exist := self.nameIndex[endpoint.Name]; exist {
		// already exist
		log.Warnf("endpoint %s already exist", endpoint.Name)
		return false
	}
	self.nameIndex[endpoint.Name] = endpoint

	if eps, ok := self.chainIndex[endpoint.Chain]; ok {
		eps.items = append(eps.items, endpoint)
	} else {
		eps := new(EndpointSet)
		eps.items = make([]*Endpoint, 1)
		eps.items[0] = endpoint
		self.chainIndex[endpoint.Chain] = eps
	}
	return true
}

func (self *Multiplexer) Select(chain ChainRef, method string) (*Endpoint, bool) {
	if eps, ok := self.chainIndex[chain]; ok {
		for i := 0; i < len(eps.items); i++ {
			idx := eps.cursor % len(eps.items)
			eps.cursor += 1

			ep := eps.items[idx]
			if !ep.Healthy {
				continue
			}

			if method != "" && ep.SkipMethods != nil {
				if _, ok := ep.SkipMethods[method]; ok {
					// the method is not provided by the endpoint, so skip it
					continue
				}
			}

			return ep, true
		}
	}
	return nil, false
}

func (self *Multiplexer) SelectOverHeight(chain ChainRef, method string, heightSpec int) (*Endpoint, bool) {

	if eps, ok := self.chainIndex[chain]; ok {
		for i := 0; i < len(eps.items); i++ {
			idx := eps.cursor % len(eps.items)
			eps.cursor += 1

			ep := eps.items[idx]

			height := heightSpec
			if heightSpec < 0 {
				height = eps.maxTipHeight + heightSpec
			}
			if !ep.Healthy {
				continue
			}

			if height >= 0 {
				if ep.Tip == nil || ep.Tip.Height < height {
					continue
				}
			}

			if method != "" && ep.SkipMethods != nil {
				if _, ok := ep.SkipMethods[method]; ok {
					// the method is not provided by the endpoint, so skip it
					continue
				}
			}

			return ep, true
		}
	}
	return nil, false
}

func (self *Multiplexer) SelectWebsocketEndpoint(chain ChainRef, method string, heightSpec int) (ep1 *Endpoint, found bool) {

	if eps, ok := self.chainIndex[chain]; ok {
		for i := 0; i < len(eps.items); i++ {
			idx := eps.cursor % len(eps.items)
			eps.cursor += 1

			ep := eps.items[idx]

			if !ep.IsWebsocket() {
				continue
			}

			height := heightSpec
			if heightSpec < 0 {
				height = eps.maxTipHeight + heightSpec
			}
			if !ep.Healthy {
				continue
			}

			if height >= 0 {
				if ep.Tip == nil || ep.Tip.Height < height {
					continue
				}
			}

			if method != "" && ep.SkipMethods != nil {
				if _, ok := ep.SkipMethods[method]; ok {
					// the method is not provided by the endpoint, so skip it
					continue
				}
			}

			return ep, true
		}
	}
	return nil, false
}

func MultiplexerFromConfig(nbcfg *NodemuxConfig) *Multiplexer {
	b := NewMultiplexer()
	b.LoadFromConfig(nbcfg)

	if nbcfg.Store.Scheme() == "redis" {
		// sync source must be a redis URL
		log.Infof("using redis store")
		chainHub, err := NewRedisChainhub(nbcfg.Store.Url)
		if err != nil {
			panic(err)
		}
		b.chainHub = chainHub
	} else {
		log.Info("using memory store")
	}
	return b
}

func (self *Multiplexer) LoadFromConfig(nbcfg *NodemuxConfig) {
	self.cfg = nbcfg
	for name, epcfg := range nbcfg.Endpoints {

		if support, _ := GetDelegatorFactory().SupportChain(epcfg.Chain); !support {
			panic(fmt.Sprintf("chain %s not supported", epcfg.Chain))
		}

		ep := NewEndpoint(name, epcfg)
		self.Add(ep)
	}
}

func (self *Multiplexer) DefaultRelayMessage(rootCtx context.Context, chain ChainRef, reqmsg *jsonz.RequestMessage, overHeight int) (jsonz.Message, error) {
	ep, found := self.SelectOverHeight(chain, reqmsg.Method, overHeight)
	if !found {
		return jsonz.ErrMethodNotFound.ToMessage(reqmsg), nil
	}
	resmsg, err := ep.CallRPC(rootCtx, reqmsg)
	return resmsg, err
}

func (self *Multiplexer) DefaultPipeREST(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := self.SelectOverHeight(chain, path, overHeight)
	if !found {
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
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}
