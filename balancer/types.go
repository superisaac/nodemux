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

type EndpointSet struct {
	items        []*Endpoint // endpoints of the same chain
	cursor       int
	maxTipHeight int
}

type Balancer struct {
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EndpointSet

	// the function to cancel sync functions
	cancelSync func()

	blockHub BlockHub
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
	TipDelegator
	DelegateREST(ctx context.Context, b *Balancer, chain ChainRef, path string, w http.ResponseWriter, r *http.Request) error
}

type DelegatorFactory struct {
	rpcDelegators  map[string]RPCDelegator
	restDelegators map[string]RESTDelegator
}

// chain stream
type BlockStatus struct {
	EndpointName string
	Chain        ChainRef
	Block        *Block
}

type BlockHub interface {
	Sub(ch chan BlockStatus)
	Unsub(ch chan BlockStatus)
	Pub() chan BlockStatus
	Run(rootCtx context.Context) error
}
