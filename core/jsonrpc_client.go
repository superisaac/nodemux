package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
	"strings"
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

func (self *Endpoint) RPCClient() jlibhttp.Client {
	self.ensureRPCClient()
	return self.rpcClient
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jlib.RequestMessage) (jlib.Message, error) {
	//self.Connect()
	self.ensureRPCClient()
	self.incrRelayCount()
	return self.rpcClient.Call(rootCtx, reqmsg)
} // CallRPC

func (self *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jlib.RequestMessage, output interface{}) error {
	//self.Connect()
	self.ensureRPCClient()
	return self.rpcClient.UnwrapCall(rootCtx, reqmsg, output)
} // UnwrapCallRPC

func (self Endpoint) IsWebsocket() bool {
	return strings.HasPrefix(self.Config.Url, "wss://") || strings.HasPrefix(self.Config.Url, "ws://")
}
