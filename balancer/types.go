package balancer

import (
	"context"
	"github.com/superisaac/jsonrpc"
	"net/http"
)

// data structures
type Block struct {
	Height int
	Hash   string
	//PrevHash string
}

type ChainRef struct {
	Name    string
	Network string
}

type AbnormalResponse struct {
	//Code int
	//Header map[string][]string
	//Body   []byte
	Response *http.Response
}

type Endpoint struct {
	// configured items
	Name          string
	Chain         ChainRef
	ServerUrl     string
	HeightPadding int
	SkipMethods   map[string]bool

	// dynamic items
	Healthy bool
	Tip     *Block

	client *http.Client
}

type EPSet struct {
	items        []*Endpoint // endpoints of the same chain
	cursor       int
	maxTipHeight int
}

type ChainDelegator interface {
	GetTip(ctx context.Context, b *Balancer, ep *Endpoint) (*Block, error)
	RequestReceived(ctx context.Context, b *Balancer, chain ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error)
}

type Balancer struct {
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EPSet

	delegators map[string]ChainDelegator

	cancelSync func()
}
