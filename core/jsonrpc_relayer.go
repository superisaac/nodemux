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

func (ep *Endpoint) ensureRPCClient() {
	if ep.rpcHttpClient == nil {
		opts := jsoffnet.ClientOptions{Timeout: ep.Config.Timeout}
		c, err := jsoffnet.NewClient(ep.Config.Url, opts)
		if err != nil {
			panic(err)
		}
		ep.rpcHttpClient = c
	}
}

func (ep *Endpoint) JSONRPCClient() jsoffnet.Client {
	ep.ensureRPCClient()
	return ep.rpcHttpClient
}

func (ep *Endpoint) NewJSONRPCWSClient() (*jsoffnet.WSClient, bool) {
	if ep.HasWebsocket() {
		c, err := jsoffnet.NewClient(ep.Config.StreamingUrl)
		if err != nil {
			panic(err)
		}
		return c.(*jsoffnet.WSClient), true
	} else {
		return nil, false
	}
}

func (ep *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsoff.RequestMessage) (jsoff.Message, error) {
	//ep.Connect()
	ep.ensureRPCClient()
	ep.incrRelayCount()

	start := time.Now()
	res, err := ep.rpcHttpClient.Call(rootCtx, reqmsg)
	// metrics the call time
	delta := time.Since(start)

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
	ep.Log().WithFields(fields).Info("call jsonrpc")
	return res, err
} // CallRPC

func (ep *Endpoint) UnwrapCallRPC(rootCtx context.Context, reqmsg *jsoff.RequestMessage, output interface{}) error {
	//ep.Connect()
	ep.ensureRPCClient()
	start := time.Now()
	err := ep.rpcHttpClient.UnwrapCall(rootCtx, reqmsg, output)

	// metrics the call time
	delta := time.Since(start)
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
	ep.Log().WithFields(fields).Info("call jsonrpc")
	return err
} // UnwrapCallRPC

func (ep Endpoint) HasWebsocket() bool {
	if ep.Config.StreamingUrl == "" {
		return false
	}
	return strings.HasPrefix(ep.Config.StreamingUrl, "wss://") || strings.HasPrefix(ep.Config.StreamingUrl, "ws://")
}
