package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jlib"
	"github.com/superisaac/jlib/http"
)

func (self *Endpoint) ensureRPCClient() {
	if self.rpcClient == nil {
		opts := jlibhttp.ClientOptions{Timeout: self.Config.Timeout}
		c, err := jlibhttp.NewClient(self.Config.Url, opts)
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

	msecs := delta.Milliseconds()

	fields := log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": msecs,
	}
	if delta.Microseconds() > 1000 {
		fields["showRequest"] = true
	}

	if err != nil {
		fields["err"] = err.Error()
	} else if res.IsError() {
		fields["err"] = fmt.Sprintf("RPC %d %s", res.MustError().Code, res.MustError().Message)
	}
	self.Log().WithFields(fields).Info("call jsonrpc")
	return res, err
} // CallRPC

func (self *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jlib.RequestMessage, output interface{}) error {
	//self.Connect()
	self.ensureRPCClient()
	start := time.Now()
	err := self.rpcClient.UnwrapCall(rootCtx, reqmsg, output)

	// metrics the call time
	delta := time.Now().Sub(start)
	fields := log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": delta.Milliseconds(),
	}
	var rpcErr *jlib.RPCError
	if err != nil {
		fields["err"] = err.Error()
	} else if errors.As(err, &rpcErr) {
		fields["err"] = fmt.Sprintf("RPCError %d %s", rpcErr.Code, rpcErr.Error())
	}
	self.Log().WithFields(fields).Info("call jsonrpc")
	return err
} // UnwrapCallRPC

func (self Endpoint) IsWebsocket() bool {
	return strings.HasPrefix(self.Config.Url, "wss://") || strings.HasPrefix(self.Config.Url, "ws://")
}
