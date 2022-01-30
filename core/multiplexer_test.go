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
		Chain:   "eosio",
		Network: "mainnet",
		Url:     "http://127.0.0.1:8899/aa/bb/",
	})
	assert.Equal("http://127.0.0.1:8899/aa/bb/v1/status", ep.FullUrl("/v1/status"))
}

func TestMultiplexer(t *testing.T) {
	assert := assert.New(t)

	b := NewMultiplexer()

	chain := ChainRef{Name: "binance-chain", Network: "mainnet"}
	ep := NewEndpoint("bsc01", EndpointConfig{
		Chain:   "binance-chain",
		Network: "mainnet",
		Url:     "http://127.0.0.1:8899",
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
