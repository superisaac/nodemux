package server

import (
	"github.com/superisaac/jsoff"
	"github.com/superisaac/jsoff/net"
	"github.com/superisaac/nodemux/core"
	"sort"
)

type rpcresultInfo struct {
	Endpoint  string `json:"endpoint"`
	URLDigest string `json:"urldigest"`
	Response  any    `json:"response,omitempty"`
	Error     string `json:"error,omitempty"`
}

func NewAdminHandler() *jsoffnet.Http1Handler {
	actor := jsoffnet.NewActor()
	actor.OnTyped("nodemux_listEndpoints", func(request *jsoffnet.RPCRequest) ([]nodemuxcore.EndpointInfo, error) {
		m := nodemuxcore.GetMultiplexer()
		infos := m.ListEndpointInfos()

		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Name < infos[j].Name
		})

		return infos, nil
	})

	actor.OnTyped("nodemux_callAll", func(request *jsoffnet.RPCRequest, chainRepr string, method string, params any) ([]rpcresultInfo, error) {
		m := nodemuxcore.GetMultiplexer()

		chain, err := nodemuxcore.ParseChain(chainRepr)
		if err != nil {
			return nil, err
		}
		reqmsg := jsoff.NewRequestMessage(1, method, params)
		results := m.BroadcastRPC(request.Context(), chain, reqmsg, -10)

		resInfos := make([]rpcresultInfo, 0)
		for _, res := range results {
			info := rpcresultInfo{
				Endpoint:  res.Endpoint.Name,
				URLDigest: res.Endpoint.URLDigest,
			}

			if res.Response != nil {
				info.Response = res.Response.Interface()
			} else {
				info.Response = nil
			}
			if res.Err != nil {
				info.Error = res.Err.Error()
			}
			resInfos = append(resInfos, info)
		}
		return resInfos, nil
	})

	return jsoffnet.NewHttp1Handler(actor)
}
