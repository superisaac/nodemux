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

func (self *Multiplexer) getBlockhead(rootCtx context.Context, ep *Endpoint, lastBlock *Block) (*Block, error) {
	logger := ep.Log()
	delegator := GetDelegatorFactory().GetBlockheadDelegator(ep.Chain.Namespace)
	block, err := delegator.GetBlockhead(rootCtx, self, ep)
	ep.incrBlockheadCount()

	if err != nil {
		logger.Warnf("mark unhealthy due to block head height error %s", err)
		ep.connected = false
		bs := ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Healthy:      false,
			Blockhead:    nil,
		}
		self.chainHub.Pub() <- bs
		return nil, err
	}
	if block != nil {
		ep.connected = true
		if !blockIsEqual(lastBlock, block) {
			self.UpdateBlock(ep, block)
		}
	} else {
		logger.Warnf("got nil head block when accessing %s %s", ep.Name, ep.Config.Url)
	}
	return block, nil
}

func (self *Multiplexer) UpdateBlock(ep *Endpoint, block *Block) {
	bs := ChainStatus{
		EndpointName: ep.Name,
		Chain:        ep.Chain,
		Healthy:      true,
		Blockhead:    block,
	}
	self.chainHub.Pub() <- bs
}

func (self *Multiplexer) UpdateBlockIfChanged(ep *Endpoint, block *Block) {
	if !blockIsEqual(ep.Blockhead, block) {
		self.UpdateBlock(ep, block)
	}
}

func (self *Multiplexer) syncEndpoint(rootCtx context.Context, ep *Endpoint) {
	delegator := GetDelegatorFactory().GetBlockheadDelegator(ep.Chain.Namespace)
	started, err := delegator.StartSync(rootCtx, self, ep)
	if err != nil {
		panic(err)
	}
	if !started {
		ep.Log().Infof("sync job not started")
		return
	}

	ep.Log().Info("sync job started")
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()
	var lastBlock *Block
	for {
		if !self.Syncing() {
			break
		}

		sleepTime := time.Duration(ep.Config.FetchInterval) * time.Second
		blk, err := self.getBlockhead(ctx, ep, lastBlock)
		if err != nil {
			// unhealthy
			ep.Log().Warnf("get block head error %s, sleep %d secs", err, 2*ep.Config.FetchInterval)
			sleepTime = time.Duration(2*ep.Config.FetchInterval) * time.Second
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
