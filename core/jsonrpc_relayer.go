package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"strings"
	"time"
)

func (self *Endpoint) ensureRPCClient() {
	if self.rpcClient == nil {
		c, err := jlibhttp.NewClient(self.Config.Url)
		if err != nil {
			panic(err)
		}
		self.rpcClient = c
	}
}

func (self *Endpoint) JSONRPCRelayer() jlibhttp.Client {
	self.ensureRPCClient()
	return self.rpcClient
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jlib.RequestMessage) (jlib.Message, error) {
	//self.Connect()
	self.ensureRPCClient()
	self.incrRelayCount()

	start := time.Now()
	res, err := self.rpcClient.Call(rootCtx, reqmsg)
	// metrics the call time
	delta := time.Now().Sub(start)
	self.Log().WithFields(log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": delta.Milliseconds(),
	}).Info("relay jsonrpc")
	return res, err
} // CallRPC

func (self *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jlib.RequestMessage, output interface{}) error {
	//self.Connect()
	self.ensureRPCClient()
	start := time.Now()
	err := self.rpcClient.UnwrapCall(rootCtx, reqmsg, output)
	// metrics the call time
	delta := time.Now().Sub(start)
	self.Log().WithFields(log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": delta.Milliseconds(),
	}).Info("relay jsonrpc")
	return err
} // UnwrapCallRPC

func (self Endpoint) IsWebsocket() bool {
	return strings.HasPrefix(self.Config.Url, "wss://") || strings.HasPrefix(self.Config.Url, "ws://")
}
