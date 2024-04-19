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
func (epset *EndpointSet) resetMaxTipHeight() {
	maxHeight := 0
	for _, epItem := range epset.items {
		if epItem.Blockhead != nil && epItem.Blockhead.Height > maxHeight {
			maxHeight = epItem.Blockhead.Height
		}
	}
	epset.maxTipHeight = maxHeight
}

func (epset EndpointSet) Get(epName string) (*Endpoint, bool) {
	ep, ok := epset.items[epName]
	return ep, ok
}

func (epset EndpointSet) MustGet(epName string) *Endpoint {
	if ep, ok := epset.Get(epName); ok {
		return ep
	}
	log.Panicf("fail to get endpoint %s", epName)
	return nil
}

func (epset EndpointSet) prometheusLabels(chain ChainRef) prometheus.Labels {
	return prometheus.Labels{
		"chain": chain.String(),
	}
}

func (epset *EndpointSet) Add(endpoint *Endpoint) {
	//epset.items = append(epset.items, endpoint)
	epset.items[endpoint.Name] = endpoint
	epset.appendWeights(endpoint)
}

func (epset *EndpointSet) appendWeights(endpoint *Endpoint) {
	if !endpoint.Healthy {
		return
	}

	w := endpoint.Config.Weight
	if w <= 0 {
		// 100 is the default weight
		w = 100
	}
	if len(epset.weights) > 0 {
		w = epset.weights[len(epset.weights)-1].AggregateValue + w
	}
	epset.weights = append(epset.weights, Weight{
		EpName:         endpoint.Name,
		AggregateValue: w,
	})
}

func (epset *EndpointSet) resetWeights() {
	weights := []Weight{}
	for _, ep := range epset.items {
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
	epset.weights = weights
}

func (epset EndpointSet) WeightLimit() int {
	if len(epset.weights) > 0 {
		return epset.weights[len(epset.weights)-1].AggregateValue
	} else {
		return 0
	}
}

func (epset EndpointSet) WeightedRandom() (string, bool) {
	if lim := epset.WeightLimit(); lim > 0 {
		w := rand.Intn(lim)
		return epset.WeightSearch(w)
	}
	return "", false
}

func (epset EndpointSet) WeightSearch(w int) (string, bool) {
	if w < 0 {
		return "", false
	}
	selected := sort.Search(len(epset.weights), func(i int) bool {
		return epset.weights[i].AggregateValue > w
	})

	if selected < len(epset.weights) {
		return epset.weights[selected].EpName, true
	} else {
		return "", false
	}
}
