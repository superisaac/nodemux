package balancer

import (
	"context"
	"time"
)

func blockIsEqual(a, b *Block) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Height == b.Height && a.Hash == b.Hash
}

func (self *Balancer) fetchTip(rootCtx context.Context, ep *Endpoint, lastBlock *Block) (*Block, error) {
	logger := ep.Log()
	delegator := GetDelegatorFactory().GetTipDelegator(ep.Chain.Name)
	block, err := delegator.GetTip(rootCtx, self, ep)
	if err != nil {
		logger.Warnf("mark unhealthy due to tip height error %s", err)
		ep.connected = false
		bs := ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Tip:          nil,
			Healthy:      false,
		}
		self.chainHub.Pub() <- bs
		return nil, err
	}
	if block != nil {
		ep.connected = true
		if !blockIsEqual(lastBlock, block) {
			bs := ChainStatus{
				EndpointName: ep.Name,
				Chain:        ep.Chain,
				Tip:          block,
				Healthy:      true,
			}
			self.chainHub.Pub() <- bs
		}
	} else {
		logger.Warnf("got nil tip block when accessing %s %s", ep.Name, ep.ServerUrl)
	}
	return block, nil
}

func (self *Balancer) fetchEndpoint(rootCtx context.Context, ep *Endpoint) {
	ep.Log().Info("fetch job started")
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()
	var lastBlock *Block
	for {
		if !self.Syncing() {
			break
		}
		sleepTime := 1 * time.Second
		blk, err := self.fetchTip(ctx, ep, lastBlock)
		if err != nil {
			// unhealthy
			sleepTime = 15 * time.Second
		}
		lastBlock = blk
		select {
		case <-ctx.Done():
			break
		case <-time.After(sleepTime):
			break
		}
	}
	ep.Log().Info("fetch job stopped")
}
