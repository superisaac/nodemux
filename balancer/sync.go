package balancer

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

func (self *Balancer) syncTip(rootCtx context.Context, ep *Endpoint) error {
	logger := ep.Log()
	delegator := GetDelegatorFactory().GetTipDelegator(ep.Chain.Name)
	block, err := delegator.GetTip(rootCtx, self, ep)
	if err != nil {
		logger.Warnf("mark unhealthy due to tip height error %s", err)
		ep.Healthy = false
		return err
	}
	if block != nil {
		bs := ChainStatus{
			EndpointName: ep.Name,
			Chain:        ep.Chain,
			Tip:          block,
		}
		self.chainHub.Pub() <- bs
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

func (self *Balancer) StartSync(rootCtx context.Context, sync bool) {
	//self.syncing = true
	if self.Syncing() {
		log.Warn("sync alredy started")
		return
	}

	ctx, cancel := context.WithCancel(rootCtx)
	self.cancelSync = cancel

	// start chainhub
	go func() {
		err := self.chainHub.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// start updater
	go self.RunUpdater(ctx)

	// start syncer
	if sync {
		for _, ep := range self.nameIndex {
			go self.syncEndpoint(ctx, ep)
		}
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

// updater
func (self *Balancer) updateStatus(bs ChainStatus) error {
	ep, ok := self.nameIndex[bs.EndpointName]
	if !ok {
		return nil
	}

	logger := ep.Log()
	block := bs.Tip

	ep.Healthy = true
	heightChanged := false

	if ep.Tip != nil {
		if ep.Tip.Height > block.Height {
			logger.Warnf("new tip height %d < old tip height %d", block.Height, ep.Tip.Height)
			heightChanged = true
		} else if ep.Tip.Height == block.Height &&
			ep.Tip.Hash != block.Hash {
			logger.Warnf("tip hash changed from %s to %s", ep.Tip.Hash, block.Hash)
		}
	}
	ep.Tip = block
	metricsEndpointBlockTip.With(ep.prometheusLabels()).Set(float64(block.Height))
	if epset, ok := self.chainIndex[ep.Chain]; ok {
		if heightChanged {
			epset.ResetMaxTipHeight()
			ep.Chain.Log().Infof("height changed, max tip height set to %d", epset.maxTipHeight)
			metricsBlockTip.With(epset.prometheusLabels(ep.Chain.Name, ep.Chain.Network)).Set(float64(epset.maxTipHeight))
		} else if epset.maxTipHeight < block.Height {
			epset.maxTipHeight = block.Height
			ep.Chain.Log().Infof("max tip height set to %d %s", epset.maxTipHeight, block.Hash)
			metricsBlockTip.With(epset.prometheusLabels(ep.Chain.Name, ep.Chain.Network)).Set(float64(epset.maxTipHeight))
		}
	} else {
		logger.Panicf("cnnot get epset by chain %s", ep.Chain)
	}
	return nil
}

func (self *Balancer) RunUpdater(rootCtx context.Context) {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	upd := make(chan ChainStatus)
	self.chainHub.Sub(upd)
	defer self.chainHub.Unsub(upd)

	for {
		select {
		case <-ctx.Done():
			return
		case bs, ok := <-upd:
			if !ok {
				return
			}
			self.updateStatus(bs)
		}
	}
}
