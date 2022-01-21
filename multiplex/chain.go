package multiplex

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

func (self ChainRef) String() string {
	return fmt.Sprintf("%s/%s", self.Name, self.Network)
}

func (self ChainRef) Empty() bool {
	return self.Name == "" || self.Network == ""
}

func (self ChainRef) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain":   self.Name,
		"network": self.Network,
	})
}
