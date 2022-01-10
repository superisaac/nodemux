package balancer

import (
	"context"
)

// implements ChainStatusHub
type LocalChainhub struct {
	pub  chan ChainStatus
	subs []chan ChainStatus
}

func NewLocalChainhub() *LocalChainhub {
	return &LocalChainhub{
		pub:  make(chan ChainStatus, 100),
		subs: make([]chan ChainStatus, 0),
	}
}

func (self *LocalChainhub) Sub(ch chan ChainStatus) {
	self.subs = append(self.subs, ch)
}

func (self *LocalChainhub) Unsub(ch chan ChainStatus) {
	found := -1
	for i, sub := range self.subs {
		if sub == ch {
			found = i
			break
		}
	}
	if found >= 0 {
		self.subs = append(self.subs[:found], self.subs[found+1:]...)
	}
}

func (self LocalChainhub) Pub() chan ChainStatus {
	return self.pub
}

func (self *LocalChainhub) Run(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// defer func() {
	// release all subs
	//self.subs = make([]chan ChainStatus, 0)
	//}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case chainSt, ok := <-self.pub:
			if !ok {
				return nil
			}
			for _, sub := range self.subs {
				sub <- chainSt
			}
		}
	}
	return nil
}
