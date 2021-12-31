package balancer

import (
	"context"
	"time"
	//log "github.com/sirupsen/logrus"
)

func (self *Balancer) syncTip(rootCtx context.Context, ep *Endpoint) error {
	logger := ep.Log()
	adaptor := self.GetAdaptor(ep.Chain.Name)
	block, err := adaptor.GetTip(rootCtx, ep)
	if err != nil {
		logger.Warnf("mark unhealthy due to tip height error %+v", err)
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
			logger.Panicf("cnnot get epset by chain %+v", ep.Chain)
		}
	} else {
		logger.Warnf("got nil tip block when accessing %s %s", ep.Name, ep.ServerUrl)
	}
	return nil
}

func (self *Balancer) syncEndpoint(rootCtx context.Context, ep *Endpoint) {
	for {
		if !self.syncing {
			break
		}
		err := self.syncTip(rootCtx, ep)
		if err != nil {
			// unhealthy
			time.Sleep(15 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (self *Balancer) StartSync(rootCtx context.Context) {
	self.syncing = true
	for _, ep := range self.nameIndex {
		go self.syncEndpoint(rootCtx, ep)
	}
}

func (self *Balancer) StopSync() {
	self.syncing = false
}
