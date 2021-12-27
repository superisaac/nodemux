package balancer

import (
//"github.com/superisaac/jsonrpc"
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
	Healthy     bool
	LatestBlock Block
}

type EPSet struct {
	items  []*Endpoint
	cursor int
}

type Balancer struct {
	// indexes
	// the name -> Endpoint map, the primary key
	nameIndex map[string]*Endpoint
	// the chain -> name map, the secondary index
	chainIndex map[ChainRef]*EPSet
}
