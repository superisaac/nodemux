package server

import (
	"github.com/superisaac/jsoff/net"
	"github.com/superisaac/nodemux/core"
	"sort"
)

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

	return jsoffnet.NewHttp1Handler(actor)
}
