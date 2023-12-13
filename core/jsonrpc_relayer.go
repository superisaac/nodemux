package nodemuxcore

// JSONRPC client from http or websocket
import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff"
	"github.com/superisaac/jsoff/net"
)

func (self *Endpoint) ensureRPCClient() {
	if self.rpcHttpClient == nil {
		opts := jsoffnet.ClientOptions{Timeout: self.Config.Timeout}
		c, err := jsoffnet.NewClient(self.Config.Url, opts)
		if err != nil {
			panic(err)
		}
		self.rpcHttpClient = c
	}
}

func (self *Endpoint) JSONRPCClient() jsoffnet.Client {
	self.ensureRPCClient()
	return self.rpcHttpClient
}

func (self *Endpoint) NewJSONRPCWSClient() (*jsoffnet.WSClient, bool) {
	if self.HasWebsocket() {
		c, err := jsoffnet.NewClient(self.Config.StreamingUrl)
		if err != nil {
			panic(err)
		}
		return c.(*jsoffnet.WSClient), true
	} else {
		return nil, false
	}
}

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsoff.RequestMessage) (jsoff.Message, error) {
	//self.Connect()
	self.ensureRPCClient()
	self.incrRelayCount()

	start := time.Now()
	res, err := self.rpcHttpClient.Call(rootCtx, reqmsg)
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

func (self *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jsoff.RequestMessage, output interface{}) error {
	//self.Connect()
	self.ensureRPCClient()
	start := time.Now()
	err := self.rpcHttpClient.UnwrapCall(rootCtx, reqmsg, output)

	// metrics the call time
	delta := time.Now().Sub(start)
	fields := log.Fields{
		"method":      reqmsg.Method,
		"timeSpentMS": delta.Milliseconds(),
	}
	var rpcErr *jsoff.RPCError
	if err != nil {
		fields["err"] = err.Error()
	} else if errors.As(err, &rpcErr) {
		fields["err"] = fmt.Sprintf("RPCError %d %s", rpcErr.Code, rpcErr.Error())
	}
	self.Log().WithFields(fields).Info("call jsonrpc")
	return err
} // UnwrapCallRPC

func (self Endpoint) HasWebsocket() bool {
	if self.Config.StreamingUrl == "" {
		return false
	}
	return strings.HasPrefix(self.Config.StreamingUrl, "wss://") || strings.HasPrefix(self.Config.StreamingUrl, "ws://")
}
