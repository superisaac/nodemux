package balancer

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
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
	b.delegators = make(map[string]ChainDelegator)
	b.Reset()
	return b
}

func (self *Balancer) Reset() {
	self.nameIndex = make(map[string]*Endpoint)
	self.chainIndex = make(map[ChainRef]*EPSet)
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
		eps := new(EPSet)
		eps.items = make([]*Endpoint, 1)
		eps.items[0] = endpoint
		self.chainIndex[endpoint.Chain] = eps
	}
	return true
}

func (self *Balancer) Select(chain ChainRef, minHeight int, method string) (*Endpoint, bool) {

	if eps, ok := self.chainIndex[chain]; ok {
		for i := 0; i < len(eps.items); i++ {
			idx := eps.cursor % len(eps.items)
			eps.cursor += 1

			ep := eps.items[idx]

			height := minHeight
			if height < 0 {
				height = eps.maxTipHeight - ep.HeightPadding
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

func (self *Balancer) LoadFromConfig(config *Config) {
	for name, epcfg := range config.Endpoints {
		ep := NewEndpoint()
		chain := ChainRef{Name: epcfg.Chain, Network: epcfg.Network}
		ep.Name = name
		ep.Chain = chain
		ep.ServerUrl = epcfg.Url
		ep.HeightPadding = epcfg.HeightPadding
		if epcfg.SkipMethods != nil {
			ep.SkipMethods = make(map[string]bool)
			for _, meth := range epcfg.SkipMethods {
				ep.SkipMethods[meth] = true
			}
		}
		self.Add(ep)
	}
}

// ChainDelegators
func (self *Balancer) Register(delegator ChainDelegator, chains ...string) {
	for _, chain := range chains {
		self.delegators[chain] = delegator
	}
}

func (self Balancer) GetDelegator(chain string) ChainDelegator {
	if delegator, ok := self.delegators[chain]; ok {
		return delegator
	}
	log.Panicf("chain %s not supported", chain)
	return nil
}

func (self *Balancer) RelayMessage(rootCtx context.Context, chain ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	ep, found := self.Select(chain, -1, reqmsg.Method)
	if !found {
		return jsonrpc.ErrMethodNotFound.ToMessage(reqmsg), nil
	}
	resmsg, err := ep.CallHTTP(rootCtx, reqmsg)
	return resmsg, err
}
