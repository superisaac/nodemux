package nodemuxcore

import (
	"context"
	log "github.com/sirupsen/logrus"
	//"time"
)

func (self Multiplexer) Syncing() bool {
	return self.cancelSync != nil
}

func (self *Multiplexer) StartSync(rootCtx context.Context, fetch bool) {
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
	go self.RunUpdator(ctx)

	// get client version
	for _, ep := range self.nameIndex {
		go ep.GetClientVersion(ctx)
	}

	// start syncer
	if fetch {
		for _, ep := range self.nameIndex {
			go self.fetchEndpoint(ctx, ep)
		}
	}
}

func (self *Multiplexer) StopSync() {
	//self.syncing = false
	if self.Syncing() {
		cancel := self.cancelSync
		self.cancelSync = nil
		cancel()
	}
}

// updater
func (self *Multiplexer) updateStatus(cs ChainStatus) error {
	ep, ok := self.Get(cs.EndpointName)
	if !ok {
		return nil
	}
	if ep.Chain != cs.Chain {
		ep.Log().Warnf("chain status mismatch, %#v", cs)
	}
	ep.Healthy = cs.Healthy

	var healthiness float64 = 0
	if ep.Healthy {
		healthiness = 1
	}
	metricsEndpointHealthy.With(ep.prometheusLabels()).Set(healthiness)
	logger := ep.Log()
	block := cs.Tip
	if block == nil {
		return nil
	}

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

func (self *Multiplexer) RunUpdator(rootCtx context.Context) {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	upd := make(chan ChainStatus, 1000)
	self.chainHub.Sub(upd)
	defer self.chainHub.Unsub(upd)

	for {
		select {
		case <-ctx.Done():
			return
		case cs, ok := <-upd:
			if !ok {
				return
			}
			self.updateStatus(cs)
		}
	}
}