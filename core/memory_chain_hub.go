package nodemuxcore

import (
	"context"
)

// implements ChainStatusHub

type MemoryChainhub struct {
	pub  chan ChainStatus
	subs []chan ChainStatus

	cmdSub   chan ChCmdChainStatus
	cmdUnsub chan ChCmdChainStatus

	snapshots map[ChainRef]ChainStatus
}

func NewMemoryChainhub() *MemoryChainhub {
	return &MemoryChainhub{
		pub:       make(chan ChainStatus, 100),
		subs:      make([]chan ChainStatus, 0),
		cmdSub:    make(chan ChCmdChainStatus, 10),
		cmdUnsub:  make(chan ChCmdChainStatus, 10),
		snapshots: make(map[ChainRef]ChainStatus, 10),
	}
}

func (self *MemoryChainhub) Sub(ch chan ChainStatus) {
	self.cmdSub <- ChCmdChainStatus{Ch: ch}
}

func (self *MemoryChainhub) subscribe(ch chan ChainStatus) {
	self.subs = append(self.subs, ch)
	for _, chainSt := range self.snapshots {
		ch <- chainSt
	}
}

func (self *MemoryChainhub) Unsub(ch chan ChainStatus) {
	self.cmdUnsub <- ChCmdChainStatus{Ch: ch}
}

func (self *MemoryChainhub) unsubscribe(ch chan ChainStatus) {
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

func (self MemoryChainhub) Pub() chan ChainStatus {
	return self.pub
}

func (self *MemoryChainhub) Run(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case cmd, ok := <-self.cmdSub:
			if !ok {
				return nil
			}
			self.subscribe(cmd.Ch)
		case cmd, ok := <-self.cmdUnsub:
			if !ok {
				return nil
			}
			self.unsubscribe(cmd.Ch)
		case chainSt, ok := <-self.pub:
			if !ok {
				return nil
			}
			self.snapshots[chainSt.Chain] = chainSt
			for _, sub := range self.subs {
				sub <- chainSt
			}
		}
	}
	return nil
}
