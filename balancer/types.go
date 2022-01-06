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
	Response *http.Response
}

type Endpoint struct {
	// configured items
	Name      string
	Chain     ChainRef
	ServerUrl string
	Headers   map[string]string
	//HeightPadding int
	SkipMethods map[string]bool

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

type Balancer struct {
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EPSet

	// the function to cancel sync functions
	cancelSync func()
}

// Delegators
type TipDelegator interface {
	GetTip(ctx context.Context, b *Balancer, ep *Endpoint) (*Block, error)
}

type RPCDelegator interface {
	TipDelegator
	DelegateRPC(ctx context.Context, b *Balancer, chain ChainRef, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error)
}

type RESTDelegator interface {
	//GetTip(ctx context.Context, b *Balancer, ep *Endpoint) (*Block, error)
	TipDelegator
	DelegateREST(ctx context.Context, b *Balancer, chain ChainRef, path string, w http.ResponseWriter, r *http.Request) error
}

type DelegatorFactory struct {
	rpcDelegators  map[string]RPCDelegator
	restDelegators map[string]RESTDelegator
}
