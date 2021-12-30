package chains

import (
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
