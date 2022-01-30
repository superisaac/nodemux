package nodemuxcore

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"net/http"
)

const (
	ApiJSONRPC = iota
	ApiJSONRPCWS
	ApiREST
	ApiGraphQL
)

// data structures
type Block struct {
	Height int
	Hash   string
}

type ChainRef struct {
	Name    string `json:"name"`
	Network string `json:"network"`
}

// type AbnormalResponse struct {
// 	Response *http.Response
// }

type Endpoint struct {
	// configured items
	Config EndpointConfig
	Name   string
	Chain  ChainRef
	//ServerUrl string
	//Headers   map[string]string
	//HeightPadding int
	SkipMethods map[string]bool

	// dynamic items
	Healthy bool
	Tip     *Block

	connected bool
	client    *http.Client
	rpcClient jsonzhttp.Client
}

type EndpointSet struct {
	items        []*Endpoint // endpoints of the same chain
	cursor       int
	maxTipHeight int
}

type Multiplexer struct {
	cfg *NodemuxConfig
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EndpointSet

	// the function to cancel sync functions
	cancelSync func()

	chainHub Chainhub

	redisClient *redis.Client
}

// Delegators
type TipDelegator interface {
	GetTip(ctx context.Context, b *Multiplexer, ep *Endpoint) (*Block, error)
}

type RPCDelegator interface {
	TipDelegator
	DelegateRPC(ctx context.Context, b *Multiplexer, chain ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error)
}

type RESTDelegator interface {
	TipDelegator
	DelegateREST(ctx context.Context, b *Multiplexer, chain ChainRef, path string, w http.ResponseWriter, r *http.Request) error
}

type GraphQLDelegator interface {
	TipDelegator
	DelegateGraphQL(ctx context.Context, b *Multiplexer, chain ChainRef, path string, w http.ResponseWriter, r *http.Request) error
}

type DelegatorFactory struct {
	rpcDelegators   map[string]RPCDelegator
	restDelegators  map[string]RESTDelegator
	graphDelegators map[string]GraphQLDelegator
}

// chain stream
type ChainStatus struct {
	EndpointName string   `json:"endpoint_name"`
	Chain        ChainRef `json:"chain"`
	Tip          *Block   `json:"tip"`
	Healthy      bool     `json:"healthy"`
}

type Chainhub interface {
	Sub(ch chan ChainStatus)
	Unsub(ch chan ChainStatus)
	Pub() chan ChainStatus
	Run(rootCtx context.Context) error
}

type ChCmdChainStatus struct {
	Ch chan ChainStatus
}
