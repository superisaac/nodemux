package balancer

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

func (self *Balancer) syncTip(rootCtx context.Context, ep *Endpoint) error {
	logger := ep.Log()
	delegator := self.GetTipDelegator(ep.Chain.Name)
	block, err := delegator.GetTip(rootCtx, self, ep)
	if err != nil {
		logger.Warnf("mark unhealthy due to tip height error %s", err)
		ep.Healthy = false
		return err
	}
	if block != nil {
		ep.Healthy = true
		if ep.Tip != nil {
			if ep.Tip.Height > block.Height {
				logger.Warnf("new tip height %d < old tip height %d", block.Height, ep.Tip.Height)
			} else if ep.Tip.Height == block.Height &&
				ep.Tip.Hash != block.Hash {
				logger.Warnf("tip hash changed from %s to %s", ep.Tip.Hash, block.Hash)
			}
		}
		ep.Tip = block
		if epset, ok := self.chainIndex[ep.Chain]; ok {
			if epset.maxTipHeight < block.Height {
				epset.maxTipHeight = block.Height
				ep.Chain.Log().Infof("max tip height set to %d", epset.maxTipHeight)
			}
		} else {
			logger.Panicf("cnnot get epset by chain %s", ep.Chain)
		}
	} else {
		logger.Warnf("got nil tip block when accessing %s %s", ep.Name, ep.ServerUrl)
	}
	return nil
}

func (self *Balancer) syncEndpoint(rootCtx context.Context, ep *Endpoint) {
	ep.Log().Info("sync job started")
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()
	for {
		if !self.Syncing() {
			break
		}
		sleepTime := 1 * time.Second
		err := self.syncTip(ctx, ep)
		if err != nil {
			// unhealthy
			sleepTime = 15 * time.Second
		}
		select {
		case <-ctx.Done():
			break
		case <-time.After(sleepTime):
			break
		}
	}
	ep.Log().Info("sync job stopped")
}

func (self Balancer) Syncing() bool {
	return self.cancelSync != nil
}

func (self *Balancer) StartSync(rootCtx context.Context) {
	//self.syncing = true
	if self.Syncing() {
		log.Warn("sync alredy started")
		return
	}
	ctx, cancel := context.WithCancel(rootCtx)
	self.cancelSync = cancel
	for _, ep := range self.nameIndex {
		go self.syncEndpoint(ctx, ep)
	}
}

func (self *Balancer) StopSync() {
	//self.syncing = false
	if self.Syncing() {
		cancel := self.cancelSync
		self.cancelSync = nil
		cancel()
	}
}
