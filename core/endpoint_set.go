package nodemuxcore

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"sort"
)

func NewEndpointSet() *EndpointSet {
	return &EndpointSet{
		items:   make([]*Endpoint, 0),
		weights: make([]int, 0),
	}
}
func (self *EndpointSet) ResetMaxTipHeight() {
	maxHeight := 0
	for _, epItem := range self.items {
		if epItem.Blockhead != nil && epItem.Blockhead.Height > maxHeight {
			maxHeight = epItem.Blockhead.Height
		}
	}
	self.maxTipHeight = maxHeight
}

func (self EndpointSet) prometheusLabels(chain ChainRef) prometheus.Labels {
	return prometheus.Labels{
		"chain": chain.String(),
	}
}

func (self *EndpointSet) Add(endpoint *Endpoint) {
	weight := endpoint.Config.Weight
	if weight <= 0 {
		// 100 is the default weight
		weight = 100
	}
	if len(self.weights) > 0 {
		// last weight
		weight = self.weights[len(self.weights)-1] + weight
	}
	self.items = append(self.items, endpoint)
	self.weights = append(self.weights, weight)
}

func (self EndpointSet) WeightLimit() int {
	if len(self.weights) > 0 {
		return self.weights[len(self.weights)-1]
	} else {
		return 0
	}
}

func (self EndpointSet) WeightRandom() (int, bool) {
	w := rand.Intn(self.WeightLimit())
	return self.WeightSearch(w)
}

func (self EndpointSet) WeightSearch(w int) (int, bool) {
	if w < 0 {
		return -1, false
	}
	selected := sort.Search(len(self.weights), func(i int) bool {
		return self.weights[i] > w
	})
	if selected < len(self.weights) {
		return selected, true
	} else {
		return -1, false
	}
}
