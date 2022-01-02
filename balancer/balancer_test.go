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
	ep := NewEndpoint()
	ep.Name = "bsc01"
	ep.Chain = chain
	ep.ServerUrl = "http://127.0.0.1:8899"

	b.Add(ep)

	assert.Equal(1, len(b.nameIndex))
	assert.Equal(1, len(b.chainIndex))

	ep1, ok := b.SelectOverHeight(chain, "", -1)
	assert.True(ok)
	assert.Equal(ep.ServerUrl, ep1.ServerUrl)

}
