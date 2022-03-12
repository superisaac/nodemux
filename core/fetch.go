package nodemuxcore

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

func (self *Multiplexer) getChaintip(rootCtx context.Context, ep *Endpoint, lastBlock *Block) (*Block, error) {
	logger := ep.Log()
	delegator := GetDelegatorFactory().GetChaintipDelegator(ep.Chain.Brand)
	block, err := delegator.GetChaintip(rootCtx, self, ep)
	if err != nil {
		logger.Warnf("mark unhealthy due to tip height error %s", err)
		ep.connected = false
		bs := ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Chaintip:     nil,
			Unhealthy:    true,
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
				Chaintip:     block,
				Unhealthy:    false,
			}
			self.chainHub.Pub() <- bs
		}
	} else {
		logger.Warnf("got nil tip block when accessing %s %s", ep.Name, ep.Config.Url)
	}
	return block, nil
}

func (self *Multiplexer) fetchEndpoint(rootCtx context.Context, ep *Endpoint) {
	delegator := GetDelegatorFactory().GetChaintipDelegator(ep.Chain.Brand)
	started, err := delegator.StartFetch(rootCtx, self, ep)
	if err != nil {
		panic(err)
	}
	if !started {
		ep.Log().Infof("fetch job not started")
		return
	}

	ep.Log().Info("fetch job started")
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()
	var lastBlock *Block
	for {
		if !self.Syncing() {
			break
		}
		sleepTime := 1 * time.Second
		blk, err := self.getChaintip(ctx, ep, lastBlock)
		if err != nil {
			// unhealthy
			ep.Log().Warnf("get chaintip error %s, sleep 15 secs", err)
			sleepTime = 5 * time.Second
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
