package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	"github.com/superisaac/jsonz"
	"github.com/superisaac/jsonz/http"
	"strings"
)

func (self *Endpoint) connectRPC() {
	if self.rpcClient == nil {
		c, err := jsonzhttp.NewClient(self.Config.Url)
		if err != nil {
			panic(err)
		}
		self.rpcClient = c
	}
}

func (self *Endpoint) RPCClient() jsonzhttp.Client {
	self.connectRPC()
	return self.rpcClient
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsonz.RequestMessage) (jsonz.Message, error) {
	//self.Connect()
	self.connectRPC()
	return self.rpcClient.Call(rootCtx, reqmsg)
} // CallRPC

func (self *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jsonz.RequestMessage, output interface{}) error {
	//self.Connect()
	self.connectRPC()
	return self.rpcClient.UnwrapCall(rootCtx, reqmsg, output)
} // UnwrapCallRPC

func (self Endpoint) IsWebsocket() bool {
	return strings.HasPrefix(self.Config.Url, "wss://") || strings.HasPrefix(self.Config.Url, "ws://")
}
