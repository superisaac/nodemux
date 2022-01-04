package chains

import (
	//"fmt"
	"context"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/superisaac/nodeb/balancer"
	"net/http"
	"strconv"
)

// type tronBlockRawData struct {
// 	Number string `mapstructure,"number"`
// 	ParentHash string `mapstructure,"parentHash"`
// }

// type tronBlockHeader struct {
// 	RawData tronBlockRawData `mapstructure,"raw_data"`
// }

// type tronBlock struct {
// 	BlockHeader tronBlockHeader `mapstructure,"block_header"`
// 	BlockID string  `mapstructure,"blockID"`
// }

type TronChain struct {
}

func resolveMap(root interface{}, path ...string) (interface{}, bool) {
	v := root
	for {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, false
		}
		hop := path[0]
		path = path[1:]

		k, ok := m[hop]
		if len(path) <= 0 {
			return k, ok
		} else {
			if !ok {
				return nil, false
			}
			v = k
		}
	}
}

func NewTronChain() *TronChain {
	return &TronChain{}
}

func (self *TronChain) GetTip(context context.Context, b *balancer.Balancer, ep *balancer.Endpoint) (*balancer.Block, error) {
	res, err := ep.RequestJson(context,
		"POST",
		"/walletsolidity/getnowblock",
		nil)
	if err != nil {
		return nil, err
	}

	// parsing height
	n1, ok := resolveMap(res, "block_header", "raw_data", "number")
	if !ok {
		return nil, errors.New("fail to get height")
	}
	var number string
	err = mapstructure.Decode(n1, &number)
	if err != nil {
		return nil, errors.Wrap(err, "decode mapstruct block")
	}

	height, err := strconv.Atoi(number)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.Atoi")
	}

	h1, ok := resolveMap(res, "blockID")
	if !ok {
		return nil, errors.New("failt to get block hash")
	}
	var blockHash string
	err = mapstructure.Decode(h1, &blockHash)
	if err != nil {
		return nil, errors.Wrap(err, "decode mapstructure. hash")
	}

	block := &balancer.Block{
		Height: height,
		Hash:   blockHash,
	}
	return block, nil
}

func (self *TronChain) RequestREST(rootCtx context.Context, b *balancer.Balancer, chain balancer.ChainRef, path string, w http.ResponseWriter, r *http.Request) error {
	// Custom relay methods can be defined here
	return b.DefaultPipeREST(rootCtx, chain, path, w, r)
}
