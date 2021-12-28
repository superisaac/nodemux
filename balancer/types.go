package balancer

import (
	"context"
	"net/http"
)

type Block struct {
	Height   int
	Hash     string
	PrevHash string
}

type ChainRef struct {
	Name    string
	Network string
}

type Endpoint struct {
	// configured items
	Name        string
	Chain       ChainRef
	ServerUrl   string
	SkipMethods map[string]bool

	//
	Healthy bool
	Tip     *Block

	client *http.Client
}

type EPSet struct {
	items        []*Endpoint
	cursor       int
	maxTipHeight int
}

type ChainAdaptor interface {
	GetTip(context context.Context, ep *Endpoint) (*Block, error)
}

type Balancer struct {
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EPSet

	adaptors map[string]ChainAdaptor
}
