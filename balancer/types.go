package balancer

import (
	"context"
	"net/http"
)

// configs
type EndpointConfig struct {
	Chain         string   `yaml:"chain"`
	Network       string   `yaml:"network"`
	Url           string   `yaml:"url"`
	SkipMethods   []string `yaml:"skip_methods,omitempty"`
	HeightPadding int      `yaml:"height_padding,omitempty"`
}

type Config struct {
	Version   string                    `yaml:"version,omitempty"`
	Endpoints map[string]EndpointConfig `yaml:"endpoints"`
}

// data structures
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

	cancelSync func()
}
