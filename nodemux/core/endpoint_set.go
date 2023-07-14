package nodemuxcore

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
)

func NewEndpointSet() *EndpointSet {
	return &EndpointSet{
		items:   make(map[string]*Endpoint),
		weights: make([]Weight, 0),
	}
}
func (self *EndpointSet) resetMaxTipHeight() {
	maxHeight := 0
	for _, epItem := range self.items {
		if epItem.Blockhead != nil && epItem.Blockhead.Height > maxHeight {
			maxHeight = epItem.Blockhead.Height
		}
	}
	self.maxTipHeight = maxHeight
}

func (self EndpointSet) Get(epName string) (*Endpoint, bool) {
	ep, ok := self.items[epName]
	return ep, ok
}

func (self EndpointSet) MustGet(epName string) *Endpoint {
	if ep, ok := self.Get(epName); ok {
		return ep
	}
	log.Panicf("fail to get endpoint %s", epName)
	return nil
}

func (self EndpointSet) prometheusLabels(chain ChainRef) prometheus.Labels {
	return prometheus.Labels{
		"chain": chain.String(),
	}
}

func (self *EndpointSet) Add(endpoint *Endpoint) {
	//self.items = append(self.items, endpoint)
	self.items[endpoint.Name] = endpoint
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

func (self EndpointSet) WeightedRandom() (string, bool) {
	if lim := self.WeightLimit(); lim > 0 {
		w := rand.Intn(lim)
		return self.WeightSearch(w)
	}
	return "", false
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
