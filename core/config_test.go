package nodemuxcore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoadYamldataParsesExtraChains(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig()

	err := cfg.LoadYamldata([]byte(`
extra_chains:
  web3:
    - fantom
    - bsc
`))

	assert.NoError(err)
	assert.Equal(map[string][]string{
		"web3": {"fantom", "bsc"},
	}, cfg.ExtraChains)
}
