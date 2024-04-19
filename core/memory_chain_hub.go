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

func (h *MemoryChainhub) Sub(ch chan ChainStatus) {
	h.cmdSub <- ChCmdChainStatus{Ch: ch}
}

func (h *MemoryChainhub) subscribe(ch chan ChainStatus) {
	h.subs = append(h.subs, ch)
	for _, chainSt := range h.snapshots {
		ch <- chainSt
	}
}

func (h *MemoryChainhub) Unsub(ch chan ChainStatus) {
	h.cmdUnsub <- ChCmdChainStatus{Ch: ch}
}

func (h *MemoryChainhub) unsubscribe(ch chan ChainStatus) {
	found := -1
	for i, sub := range h.subs {
		if sub == ch {
			found = i
			break
		}
	}
	if found >= 0 {
		h.subs = append(h.subs[:found], h.subs[found+1:]...)
	}
}

func (h MemoryChainhub) Pub() chan ChainStatus {
	return h.pub
}

func (h *MemoryChainhub) Run(rootCtx context.Context) error {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case cmd, ok := <-h.cmdSub:
			if !ok {
				return nil
			}
			h.subscribe(cmd.Ch)
		case cmd, ok := <-h.cmdUnsub:
			if !ok {
				return nil
			}
			h.unsubscribe(cmd.Ch)
		case chainSt, ok := <-h.pub:
			if !ok {
				return nil
			}
			h.snapshots[chainSt.Chain] = chainSt
			for _, sub := range h.subs {
				sub <- chainSt
			}
		}
	}
}
