package server

import (
	"github.com/superisaac/jsonz/http"
	"github.com/superisaac/nodemux/core"
	"sort"
)

func NewAdminHandler() *jsonzhttp.H1Handler {
	actor := jsonzhttp.NewActor()
	actor.OnTyped("nodemux_listEndpoints", func(request *jsonzhttp.RPCRequest) ([]nodemuxcore.EndpointInfo, error) {
		m := nodemuxcore.GetMultiplexer()
		infos := m.ListEndpointInfos()

		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Name < infos[j].Name
		})

		return infos, nil
	})

	return jsonzhttp.NewH1Handler(actor)
}
