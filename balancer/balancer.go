package balancer

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/nodeb/cfg"
	"net/http"
	//"sync"
)

var (
	_instance *Balancer
	//once      sync.Once
)

func GetBalancer() *Balancer {
	//once.Do(func() {
	//	_instance = NewBalancer()
	//})
	if _instance == nil {
		log.Panicf("Balancer instance is nil")
	}
	return _instance
}

func SetBalancer(b *Balancer) {
	if _instance != nil {
		_instance.StopSync()
	}
	_instance = b
}

func NewBalancer() *Balancer {
	b := new(Balancer)
	b.chainHub = NewLocalChainhub()
	b.Reset()
	return b
}

func (self *Balancer) Reset() {
	self.nameIndex = make(map[string]*Endpoint)
	self.chainIndex = make(map[ChainRef]*EndpointSet)
}

func (self *Balancer) Add(endpoint *Endpoint) bool {
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

func (self *Balancer) Select(chain ChainRef, method string) (*Endpoint, bool) {
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

func (self *Balancer) SelectOverHeight(chain ChainRef, method string, heightSpec int) (*Endpoint, bool) {

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

func BalancerFromConfig(nbcfg *cfg.NodebConfig) *Balancer {
	b := NewBalancer()
	b.LoadFromConfig(nbcfg)
	if nbcfg.SyncSource != "" {
		// sync source must be a redis URL
		chainHub, err := NewRedisChainhub(nbcfg.SyncSource)
		if err != nil {
			panic(err)
		}
		b.chainHub = chainHub
	}
	return b
}

func (self *Balancer) LoadFromConfig(nbcfg *cfg.NodebConfig) {
	self.cfg = nbcfg
	for name, epcfg := range nbcfg.Endpoints {
		ep := NewEndpoint()
		if !GetDelegatorFactory().SupportChain(epcfg.Chain) {
			panic(fmt.Sprintf("chain %s not supported", epcfg.Chain))
		}
		chain := ChainRef{Name: epcfg.Chain, Network: epcfg.Network}
		ep.Name = name
		ep.Chain = chain
		ep.ServerUrl = epcfg.Url
		ep.Headers = epcfg.Headers
		if epcfg.SkipMethods != nil {
			ep.SkipMethods = make(map[string]bool)
			for _, meth := range epcfg.SkipMethods {
				ep.SkipMethods[meth] = true
			}
		}
		self.Add(ep)
	}
}

func (self *Balancer) DefaultRelayMessage(rootCtx context.Context, chain ChainRef, reqmsg *jsonrpc.RequestMessage, overHeight int) (jsonrpc.IMessage, error) {
	ep, found := self.SelectOverHeight(chain, reqmsg.Method, overHeight)
	if !found {
		return jsonrpc.ErrMethodNotFound.ToMessage(reqmsg), nil
	}
	resmsg, err := ep.CallRPC(rootCtx, reqmsg)
	return resmsg, err
}

func (self *Balancer) DefaultPipeREST(rootCtx context.Context, chain ChainRef, path string, w http.ResponseWriter, r *http.Request, overHeight int) error {
	ep, found := self.SelectOverHeight(chain, path, overHeight)
	if !found {
		//return jsonrpc.ErrMethodNotFound.ToMessage(reqmsg), nil
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return nil
	}
	err := ep.PipeRequest(rootCtx, path, w, r)
	return err
}
