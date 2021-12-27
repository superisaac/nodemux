package balancer

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestBalancer(t *testing.T) {
	assert := assert.New(t)

	b := NewBalancer()

	chain := ChainRef{Name: "binance-chain", Network: "mainnet"}
	ep := &Endpoint{Name: "binance-chain", Chain: chain, ServerUrl: "http://127.0.0.1:5432", Healthy: true}

	b.Add(ep)

	assert.Equal(1, len(b.nameIndex))
	assert.Equal(1, len(b.chainIndex))

	ep1, ok := b.Select(chain, -1, "")
	assert.True(ok)
	assert.Equal(ep.ServerUrl, ep1.ServerUrl)
}
