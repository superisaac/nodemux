package multiplex

// JSONRPC client from http or websocket
import (
	"context"
	"github.com/superisaac/jsonrpc"
	"github.com/superisaac/jsonrpc/http"
)

func (self *Endpoint) connectRPC() {
	if self.rpcClient == nil {
		c, err := jsonrpchttp.GetClient(self.ServerUrl)
		if err != nil {
			panic(err)
		}
		self.rpcClient = c
	}
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	//self.Connect()
	self.connectRPC()
	return self.rpcClient.Call(rootCtx, reqmsg)
} // CallRPC
