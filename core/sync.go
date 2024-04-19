package nodemuxcore

import (
	"context"
	log "github.com/sirupsen/logrus"
	//"time"
)

func (m Multiplexer) Syncing() bool {
	return m.cancelSync != nil
}

func (m *Multiplexer) StartSync(rootCtx context.Context, fetch bool) {
	//m.syncing = true
	if m.Syncing() {
		log.Warn("sync alredy started")
		return
	}

	ctx, cancel := context.WithCancel(rootCtx)
	m.cancelSync = cancel

	// start chainhub
	go func() {
		err := m.chainHub.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	// start updater
	go m.RunUpdator(ctx)

	// get client version
	for _, ep := range m.nameIndex {
		go ep.GetClientVersion(ctx)
	}

	// start syncer
	if fetch {
		for _, ep := range m.nameIndex {
			go m.syncEndpoint(ctx, ep)
		}
	}
}

func (m *Multiplexer) StopSync() {
	//m.syncing = false
	if m.Syncing() {
		cancel := m.cancelSync
		m.cancelSync = nil
		cancel()
	}
}

// updater
func (m *Multiplexer) updateStatus(cs ChainStatus) error {
	ep, ok := m.Get(cs.EndpointName)
	if !ok {
		return nil
	}
	logger := ep.Log()
	if ep.Chain != cs.Chain {
		logger.Warnf("chain status mismatch, %#v", cs)
	}
	if cs.Healthy != ep.Healthy {
		ep.Healthy = cs.Healthy
		logger.Infof("healthy set to %t", cs.Healthy)
		eps := m.chainIndex[ep.Chain]
		eps.resetWeights()
	}

	var healthy float64 = 0
	if ep.Healthy {
		healthy = 1
	}
	metricsEndpointHealthy.With(ep.prometheusLabels()).Set(healthy)

	block := cs.Blockhead
	if block == nil {
		return nil
	}

	heightChanged := false

	if ep.Blockhead != nil {
		if ep.Blockhead.Height > block.Height {
			logger.Warnf("new block head height %d < old block head height %d",
				block.Height,
				ep.Blockhead.Height)
			heightChanged = true
		} else if ep.Blockhead.Height == block.Height &&
			ep.Blockhead.Hash != block.Hash {
			logger.Warnf("block head hash changed from %s to %s",
				ep.Blockhead.Hash,
				block.Hash)
		}
	}
	ep.Blockhead = block

	metricsEndpointBlockTip.With(
		ep.prometheusLabels()).Set(
		float64(block.Height))

	if epset, ok := m.chainIndex[ep.Chain]; ok {
		if heightChanged {
			epset.resetMaxTipHeight()
			ep.Chain.Log().Infof(
				"height changed, max block head height set to %d",
				epset.maxTipHeight)
			metricsBlockTip.With(
				epset.prometheusLabels(ep.Chain)).Set(
				float64(epset.maxTipHeight))
		} else if epset.maxTipHeight < block.Height {
			epset.maxTipHeight = block.Height
			ep.Chain.Log().Infof(
				"max block head height set to %d %s",
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

func (m *Multiplexer) RunUpdator(rootCtx context.Context) {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	upd := make(chan ChainStatus, 1000)
	m.chainHub.Sub(upd)
	defer m.chainHub.Unsub(upd)

	for {
		select {
		case <-ctx.Done():
			return
		case cs, ok := <-upd:
			if !ok {
				return
			}
			m.updateStatus(cs)
		}
	}
}
