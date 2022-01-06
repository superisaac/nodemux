package balancer

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	_factory *DelegatorFactory
	once     sync.Once
)

func GetDelegatorFactory() *DelegatorFactory {
	once.Do(func() {
		_factory = newDelegatorFactory()
	})
	return _factory
}

func newDelegatorFactory() *DelegatorFactory {
	return &DelegatorFactory{
		rpcDelegators:  make(map[string]RPCDelegator),
		restDelegators: make(map[string]RESTDelegator),
	}
}

func (self DelegatorFactory) SupportChain(chain string) bool {
	if _, ok := self.rpcDelegators[chain]; ok {
		return true
	}
	if _, ok := self.restDelegators[chain]; ok {
		return true
	}
	return false
}

func (self DelegatorFactory) GetTipDelegator(chain string) TipDelegator {
	if delg, ok := self.rpcDelegators[chain]; ok {
		return delg
	} else if delg, ok := self.restDelegators[chain]; ok {
		return delg
	}
	log.Panicf("chain %s not supported", chain)
	return nil
}

// RPC delegators
func (self *DelegatorFactory) RegisterRPC(delegator RPCDelegator, chains ...string) {
	for _, chain := range chains {
		self.rpcDelegators[chain] = delegator
	}
}

func (self DelegatorFactory) GetRPCDelegator(chain string) RPCDelegator {
	if delegator, ok := self.rpcDelegators[chain]; ok {
		return delegator
	}
	log.Panicf("chain %s not supported", chain)
	return nil
}

// REST delegators
func (self *DelegatorFactory) RegisterREST(delegator RESTDelegator, chains ...string) {
	for _, chain := range chains {
		self.restDelegators[chain] = delegator
	}
}

func (self DelegatorFactory) GetRESTDelegator(chain string) RESTDelegator {
	if delegator, ok := self.restDelegators[chain]; ok {
		return delegator
	}
	log.Panicf("chain %s not supported", chain)
	return nil
}
