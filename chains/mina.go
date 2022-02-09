package chains

import (
	"context"
	"github.com/superisaac/nodemux/core"
	"net/http"
)

type minaBlock struct {
	StateHash     string
	ProtocolState struct {
		ConsensusState struct {
			BlockHeight int
		}
	}
}

type minaTipResult struct {
	BestChain []minaBlock
}

type MinaChain struct {
}

func NewMinaChain() *MinaChain {
	return &MinaChain{}
}

func (self MinaChain) GetClientVersion(context context.Context, ep *nodemuxcore.Endpoint) (string, error) {
	return "", nil
}

func (self *MinaChain) GetChaintip(context context.Context, b *nodemuxcore.Multiplexer, ep *nodemuxcore.Endpoint) (*nodemuxcore.Block, error) {
	q := `{
  bestChain(maxLength: 1){
    stateHash
    protocolState {
      consensusState {
        blockHeight
      }
    }
  }
}`

	var res minaTipResult
	err := ep.RequestGraphQL(context, q, nil, nil, &res)
	if err != nil {
		return nil, err
	}

	if len(res.BestChain) > 0 {
		blk := res.BestChain[0]
		block := &nodemuxcore.Block{
			Height: blk.ProtocolState.ConsensusState.BlockHeight,
			Hash:   blk.StateHash,
		}
		return block, nil
	} else {
		ep.Log().Info("query does not return a block")
		return nil, nil
	}

}

func (self *MinaChain) DelegateGraphQL(rootCtx context.Context, b *nodemuxcore.Multiplexer, chain nodemuxcore.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeGraphQL(rootCtx, chain, path, w, r, -10)
}
