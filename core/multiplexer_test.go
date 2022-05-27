package nodemuxcore

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestEndpoint(t *testing.T) {
	assert := assert.New(t)

	ep := NewEndpoint("tron01", EndpointConfig{
		Chain: "eosio/mainnet",
		Url:   "http://127.0.0.1:8899/aa/bb/",
	})
	assert.Equal("http://127.0.0.1:8899/aa/bb/v1/status", ep.FullUrl("/v1/status"))
}

func TestWeight(t *testing.T) {
	assert := assert.New(t)

	ep1 := NewEndpoint("tron01", EndpointConfig{
		Chain:  "eosio/mainnet",
		Url:    "http://127.0.0.1:8899/aa/bb/",
		Weight: 150, // weight range [0, 150)
	})

	ep2 := NewEndpoint("tron02", EndpointConfig{
		Chain:  "eosio/mainnet",
		Url:    "http://127.0.0.1:8899/aa/bb/",
		Weight: 60, // weight range [150, 210)
	})

	eps := NewEndpointSet()
	eps.Add(ep1)
	eps.Add(ep2)

	assert.Equal(210, eps.WeightLimit())

	epName, ok := eps.WeightSearch(0)
	assert.True(ok)
	assert.Equal("tron01", epName)

	epName, ok = eps.WeightSearch(70)
	assert.True(ok)
	assert.Equal("tron01", epName)

	epName, ok = eps.WeightSearch(150)
	assert.True(ok)
	assert.Equal("tron02", epName)

	epName, ok = eps.WeightSearch(151)
	assert.True(ok)
	assert.Equal("tron02", epName)

	epName, ok = eps.WeightSearch(210)
	assert.False(ok)

	epName, ok = eps.WeightSearch(211)
	assert.False(ok)

	epName, ok = eps.WeightSearch(211)
	assert.False(ok)

	epName, ok = eps.WeightSearch(-3)
	assert.False(ok)
}

func TestMultiplexer(t *testing.T) {
	assert := assert.New(t)

	b := NewMultiplexer()

	chain := ChainRef{
		Brand:   "binance-chain",
		Network: "mainnet",
	}
	ep := NewEndpoint("bsc01", EndpointConfig{
		Chain: "binance-chain/mainnet",
		Url:   "http://127.0.0.1:8899",
	})
	b.Add(ep)

	assert.Equal(1, len(b.nameIndex))
	assert.Equal(1, len(b.chainIndex))

	ep1, ok := b.SelectOverHeight(chain, "", -1)
	assert.True(ok)
	assert.Equal(ep.Config.Url, ep1.Config.Url)
}

func TestUrlParse(t *testing.T) {
	assert := assert.New(t)

	u, err := url.Parse("memory")
	assert.Nil(err)
	assert.Equal("", u.Scheme)

	u, err = url.Parse("memory:")
	assert.Nil(err)
	assert.Equal("memory", u.Scheme)

	u, err = url.Parse("redis://localhost")
	assert.Nil(err)
	assert.Equal("redis", u.Scheme)
}
