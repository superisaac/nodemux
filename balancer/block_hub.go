package balancer

import (
	"context"
)

type LocalBlockHub struct {
	pub  chan BlockStatus
	subs []chan BlockStatus
}

func NewLocalBlockHub() *LocalBlockHub {
	return &LocalBlockHub{
		pub:  make(chan BlockStatus, 100),
		subs: make([]chan BlockStatus, 0),
	}
}

func (self *LocalBlockHub) Sub(ch chan BlockStatus) {
	self.subs = append(self.subs, ch)
}

func (self *LocalBlockHub) Unsub(ch chan BlockStatus) {
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

func (self LocalBlockHub) Pub() chan BlockStatus {
	return self.pub
}

func (self *LocalBlockHub) Run(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	defer func() {
		// release all subs
		//self.subs = make([]chan BlockStatus, 0)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case blockSt, ok := <-self.pub:
			if !ok {
				return nil
			}
			for _, sub := range self.subs {
				sub <- blockSt
			}
		}
	}
	return nil
}
