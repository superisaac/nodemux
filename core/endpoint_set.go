package nodemuxcore

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"sort"
)

func NewEndpointSet() *EndpointSet {
	return &EndpointSet{
		items:   make([]*Endpoint, 0),
		weights: make([]Weight, 0),
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
	self.items = append(self.items, endpoint)
	self.appendWeights(endpoint)
}

func (self *EndpointSet) appendWeights(endpoint *Endpoint) {
	if !endpoint.Healthy {
		return
	}

	w := endpoint.Config.Weight
	if w <= 0 {
		// 100 is the default weight
		w = 100
	}
	if len(self.weights) > 0 {
		w = self.weights[len(self.weights)-1].AggregateValue + w
	}
	self.weights = append(self.weights, Weight{
		EpName:         endpoint.Name,
		AggregateValue: w,
	})
}

func (self *EndpointSet) resetWeights() {
	weights := []Weight{}
	for _, ep := range self.items {
		if !ep.Healthy {
			continue
		}
		w := ep.Config.Weight
		if w <= 0 {
			// 100 is the default weight
			w = 100
		}
		if len(weights) > 0 {
			w = weights[len(weights)-1].AggregateValue + w
		}
		weights = append(weights, Weight{
			EpName:         ep.Name,
			AggregateValue: w,
		})
	}
	self.weights = weights
}

func (self EndpointSet) WeightLimit() int {
	if len(self.weights) > 0 {
		return self.weights[len(self.weights)-1].AggregateValue
	} else {
		return 0
	}
}

func (self EndpointSet) WeightRandom() (string, bool) {
	w := rand.Intn(self.WeightLimit())
	return self.WeightSearch(w)
}

func (self EndpointSet) WeightSearch(w int) (string, bool) {
	if w < 0 {
		return "", false
	}
	selected := sort.Search(len(self.weights), func(i int) bool {
		return self.weights[i].AggregateValue > w
	})

	if selected < len(self.weights) {
		return self.weights[selected].EpName, true
	} else {
		return "", false
	}
}
