package server

import (
	"github.com/superisaac/jlib/http"
	"github.com/superisaac/nodemux/core"
	"sort"
)

func NewAdminHandler() *jlibhttp.H1Handler {
	actor := jlibhttp.NewActor()
	actor.OnTyped("nodemux_listEndpoints", func(request *jlibhttp.RPCRequest) ([]nodemuxcore.EndpointInfo, error) {
		m := nodemuxcore.GetMultiplexer()
		infos := m.ListEndpointInfos()

		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Name < infos[j].Name
		})

		return infos, nil
	})

	return jlibhttp.NewH1Handler(actor)
}
