package balancer

import (
	"sync"
	//yaml "gopkg.in/yaml.v2"
)

var (
	_balancer *Balancer
	once      sync.Once
)

func GetBalancer() *Balancer {
	once.Do(func() {
		_balancer = NewBalancer()
	})
	return _balancer
}

func NewBalancer() *Balancer {
	b := new(Balancer)
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

func (self *Balancer) Select(chain ChainRef, height int, method string) (*Endpoint, bool) {

	if eps, ok := self.chainIndex[chain]; ok {
		for i := 0; i < len(eps.items); i++ {
			idx := eps.cursor % len(eps.items)
			eps.cursor += 1

			ep := eps.items[idx]
			if !ep.Healthy {
				continue
			}

			if height > 0 {
				if ep.LatestBlock.Hash == "" || ep.LatestBlock.Height < height {
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
		if epcfg.SkipMethods != nil {
			ep.SkipMethods = make(map[string]bool)
			for _, meth := range epcfg.SkipMethods {
				ep.SkipMethods[meth] = true
			}
		}
		self.Add(ep)
	}
}
