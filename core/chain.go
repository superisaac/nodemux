package nodemuxcore

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

func ParseChain(chainRepr string) (ChainRef, error) {
	arr := strings.SplitN(chainRepr, "/", 2)
	if len(arr) != 2 {
		//panic("invalid chain format")
		return ChainRef{}, errors.New("invalid chain format")
	}
	return ChainRef{
		Namespace: arr[0],
		Network:   arr[1],
	}, nil
}

func MustParseChain(chainRepr string) ChainRef {
	chain, err := ParseChain(chainRepr)
	if err != nil {
		log.Panicf("parse chain %s", err)
	}
	return chain
}

func (self ChainRef) String() string {
	return fmt.Sprintf("%s/%s", self.Namespace, self.Network)
}

func (self ChainRef) Empty() bool {
	return self.Namespace == "" || self.Network == ""
}

func (self ChainRef) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain": self.String(),
	})
}
