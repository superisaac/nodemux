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
			go self.syncEndpoint(ctx, ep)
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
	logger := ep.Log()
	if ep.Chain != cs.Chain {
		logger.Warnf("chain status mismatch, %#v", cs)
	}
	if cs.Unhealthy != ep.Unhealthy {
		ep.Unhealthy = cs.Unhealthy
		logger.Infof("unhealthy set to %t", cs.Unhealthy)
	}

	var healthiness float64 = 1
	if ep.Unhealthy {
		healthiness = 0
	}
	metricsEndpointHealthy.With(ep.prometheusLabels()).Set(healthiness)

	block := cs.Chaintip
	if block == nil {
		return nil
	}

	heightChanged := false

	if ep.Chaintip != nil {
		if ep.Chaintip.Height > block.Height {
			logger.Warnf("new tip height %d < old tip height %d",
				block.Height,
				ep.Chaintip.Height)
			heightChanged = true
		} else if ep.Chaintip.Height == block.Height &&
			ep.Chaintip.Hash != block.Hash {
			logger.Warnf("tip hash changed from %s to %s",
				ep.Chaintip.Hash,
				block.Hash)
		}
	}
	ep.Chaintip = block

	metricsEndpointBlockTip.With(
		ep.prometheusLabels()).Set(
		float64(block.Height))

	if epset, ok := self.chainIndex[ep.Chain]; ok {
		if heightChanged {
			epset.ResetMaxTipHeight()
			ep.Chain.Log().Infof(
				"height changed, max tip height set to %d",
				epset.maxTipHeight)
			metricsBlockTip.With(
				epset.prometheusLabels(ep.Chain)).Set(
				float64(epset.maxTipHeight))
		} else if epset.maxTipHeight < block.Height {
			epset.maxTipHeight = block.Height
			ep.Chain.Log().Infof(
				"max tip height set to %d %s",
				epset.maxTipHeight,
				block.Hash)
			metricsBlockTip.With(
				epset.prometheusLabels(ep.Chain)).Set(
				float64(epset.maxTipHeight))
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
