package balancer

import (
	"context"
	//log "github.com/sirupsen/logrus"
)

func (self *Balancer) SyncTip(rootCtx context.Context, ep *Endpoint) error {
	adaptor := self.GetAdaptor(ep.Chain.Name)
	block, err := adaptor.GetTip(rootCtx, ep)
	if err != nil {
		ep.Healthy = false
		return err
	}
	if block != nil {
		ep.Healthy = true
		if ep.Tip != nil {
			if ep.Tip.Height > block.Height {
				ep.Log().Warnf("new tip height %d < old tip height %d", block.Height, ep.Tip.Height)
			} else if ep.Tip.Height == block.Height &&
				ep.Tip.Hash != block.Hash {
				ep.Log().Warnf("tip hash changed from %s to %s", ep.Tip.Hash, block.Hash)
			}
		}
		ep.Tip = block
		if epset, ok := self.chainIndex[ep.Chain]; ok {
			if epset.maxTipHeight < block.Height {
				epset.maxTipHeight = block.Height
			}
		} else {
			ep.Log().Panicf("cnnot get epset by chain %+v", ep.Chain)
		}
	} else {
		ep.Log().Warnf("got nil tip block when accessing %s %s", ep.Name, ep.ServerUrl)
	}
	return nil
}
