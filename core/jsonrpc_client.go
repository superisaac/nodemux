package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	"github.com/superisaac/jsoz"
	"github.com/superisaac/jsoz/http"
)

func (self *Endpoint) connectRPC() {
	if self.rpcClient == nil {
		c, err := jsozhttp.GetClient(self.Config.Url)
		if err != nil {
			panic(err)
		}
		self.rpcClient = c
	}
}

func (self *Endpoint) RPCClient() jsozhttp.Client {
	return self.rpcClient
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsoz.RequestMessage) (jsoz.Message, error) {
	//self.Connect()
	self.connectRPC()
	return self.rpcClient.Call(rootCtx, reqmsg)
} // CallRPC
