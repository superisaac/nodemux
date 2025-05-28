package chains

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func TestEthereum(t *testing.T) {
	assert := assert.New(t)

	d, err := hexutil.DecodeUint64("0x789")
	assert.Nil(err)
	assert.Equal(uint64(1929), d)
}

func TestResolveMap(t *testing.T) {
	assert := assert.New(t)
	v := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"d": 5,
			},
		},
	}

	r, ok := resolveMap(v, "a", "b", "d")
	assert.True(ok)
	assert.Equal(5, r)
}

func TestMapTronBlock(t *testing.T) {
	assert := assert.New(t)

	body := []byte(`{
"blockID": "abc",
"block_header": {
  "raw_data": {
    "number": 123,
    "parentHash": "def"}
  }
}`)

	var v tronBlock
	err := json.Unmarshal(body, &v)
	assert.Nil(err)

	assert.Equal("abc", v.BlockID)

	assert.Equal("def", v.BlockHeader.RawData.ParentHash)
	assert.Equal(123, v.BlockHeader.RawData.Number)
}
