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
	Height int    `json:"height"`
	Hash   string `json:"hash,omitempty"`
}

type ChainRef struct {
	Brand   string
	Network string
}

type Endpoint struct {
	// configured items
	Config      EndpointConfig
	Name        string
	Chain       ChainRef
	SkipMethods map[string]bool

	// fetched
	ClientVersion string

	// dynamic items
	Healthy   bool
	Blockhead *Block

	client    *http.Client
	rpcClient jsonzhttp.Client

	// sync status
	connected bool
}

type EndpointInfo struct {
	Name          string `json:"name"`
	Chain         string `json:"chain"`
	Healthy       bool   `json:"healthy"`
	Blockhead     *Block `json:"head,omitempty"`
	ClientVersion string `json:"client,omitempty"`
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

	// the pubsub hub of chain status messages
	chainHub Chainhub

	// a pool of redis clients
	redisClients map[string]*redis.Client
}

// Delegators
type BlockheadDelegator interface {
	// if an endpoint want to custom the sync procedure then the
	// func should return false else the func returns true
	StartSync(ctx context.Context, m *Multiplexer, ep *Endpoint) (started bool, err error)

	// Get a block head
	GetBlockhead(ctx context.Context, m *Multiplexer, ep *Endpoint) (*Block, error)

	// Get the client version
	GetClientVersion(ctx context.Context, ep *Endpoint) (string, error)
}

type RPCDelegator interface {
	BlockheadDelegator
	DelegateRPC(ctx context.Context, b *Multiplexer, chain ChainRef, reqmsg *jsonz.RequestMessage) (jsonz.Message, error)
}

type RESTDelegator interface {
	BlockheadDelegator
	DelegateREST(ctx context.Context, b *Multiplexer, chain ChainRef, path string, w http.ResponseWriter, r *http.Request) error
}

type GraphQLDelegator interface {
	BlockheadDelegator
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
	Blockhead    *Block   `json:"head"`
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
