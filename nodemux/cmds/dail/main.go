package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/superisaac/nodemux/chains"
	"github.com/superisaac/nodemux/core"
	"os"
	"time"
)

func main() {
	dailFlags := flag.NewFlagSet("nodemux-dail", flag.ExitOnError)
	pChain := dailFlags.String("chain", "", "the chain namespace/network")
	pUrl := dailFlags.String("url", "", "the rpc url to connect")

	dailFlags.Parse(os.Args[1:])
	factory := nodemuxcore.GetDelegatorFactory()
	chains.InstallAdaptors(factory)

	chainref, err := nodemuxcore.ParseChain(*pChain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse chain failed, %s\n", err)
		os.Exit(1)
	}

	if *pUrl == "" {
		fmt.Fprintf(os.Stderr, "empty url\n")
		os.Exit(1)
	}
	if support, _ := factory.SupportChain(chainref.Namespace); !support {
		fmt.Fprintf(os.Stderr, "chain %s not supported \n", chainref.Namespace)
		os.Exit(1)
	}

	epcfg := nodemuxcore.EndpointConfig{
		Chain: chainref.String(),
		Url:   *pUrl,
	}
	endpoint := nodemuxcore.NewEndpoint("endpoint001", epcfg)
	multiplexer := nodemuxcore.NewMultiplexer()
	nodemuxcore.SetMultiplexer(multiplexer)

	delegator := factory.GetBlockheadDelegator(endpoint.Chain.Namespace)
	start := time.Now()
	block, err := delegator.GetBlockhead(context.Background(), multiplexer, endpoint)
	delta := time.Now().Sub(start)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got block head height as %d in %d ms\n", block.Height, delta.Milliseconds())
}
