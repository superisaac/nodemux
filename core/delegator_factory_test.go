package nodemuxcore

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/superisaac/jsoff"
)

type testDelegator struct {
	namespace string
}

func (d testDelegator) Namespace() string {
	return d.namespace
}

func (d testDelegator) StartSync(context.Context, *Multiplexer, *Endpoint) (bool, error) {
	return true, nil
}

func (d testDelegator) GetBlockhead(context.Context, *Multiplexer, *Endpoint) (*Block, error) {
	return nil, nil
}

func (d testDelegator) GetClientVersion(context.Context, *Endpoint) (string, error) {
	return "", nil
}

func (d testDelegator) DelegateRPC(context.Context, *Multiplexer, ChainRef, *jsoff.RequestMessage, *http.Request) (jsoff.Message, error) {
	return nil, nil
}

func (d testDelegator) DelegateREST(context.Context, *Multiplexer, ChainRef, string, http.ResponseWriter, *http.Request) error {
	return nil
}

func (d testDelegator) DelegateGraphQL(context.Context, *Multiplexer, ChainRef, string, http.ResponseWriter, *http.Request) error {
	return nil
}

func TestRegisterRPCAddsConfiguredChainsForDelegatorNamespace(t *testing.T) {
	assert := assert.New(t)
	factory := newDelegatorFactory()
	factory.SetConfig(&NodemuxConfig{
		ExtraChains: map[string][]string{
			"web3": {"bsc", "fantom-web3"},
		},
	})

	delegator := &testDelegator{namespace: "web3"}
	factory.RegisterRPC(delegator, "ethereum")

	assert.Same(delegator, factory.rpcDelegators["ethereum"])
	assert.Same(delegator, factory.rpcDelegators["bsc"])
	assert.Same(delegator, factory.rpcDelegators["fantom-web3"])
}

func TestRegisterRESTAddsConfiguredChainsForDelegatorNamespace(t *testing.T) {
	assert := assert.New(t)
	factory := newDelegatorFactory()
	factory.SetConfig(&NodemuxConfig{
		ExtraChains: map[string][]string{
			"tron": {"tron-grid"},
		},
	})

	delegator := &testDelegator{namespace: "tron"}
	factory.RegisterREST(delegator, "tron-full")

	assert.Same(delegator, factory.restDelegators["tron-full"])
	assert.Same(delegator, factory.restDelegators["tron-grid"])
}

func TestRegisterGraphQLAddsConfiguredChainsForDelegatorNamespace(t *testing.T) {
	assert := assert.New(t)
	factory := newDelegatorFactory()
	factory.SetConfig(&NodemuxConfig{
		ExtraChains: map[string][]string{
			"fantom": {"fantom-graphql"},
		},
	})

	delegator := &testDelegator{namespace: "fantom"}
	factory.RegisterGraphQL(delegator, "fantom")

	assert.Same(delegator, factory.graphDelegators["fantom"])
	assert.Same(delegator, factory.graphDelegators["fantom-graphql"])
}
