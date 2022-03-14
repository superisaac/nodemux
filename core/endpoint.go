package nodemuxcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonz/http"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// EndpointSet
func (self *EndpointSet) ResetMaxTipHeight() {
	maxHeight := 0
	for _, epItem := range self.items {
		if epItem.Chaintip != nil && epItem.Chaintip.Height > maxHeight {
			maxHeight = epItem.Chaintip.Height
		}
	}
	self.maxTipHeight = maxHeight
}

func (self EndpointSet) prometheusLabels(chain ChainRef) prometheus.Labels {
	return prometheus.Labels{
		"chain": chain.String(),
	}
}

/// Create an endpoint instance
func NewEndpoint(name string, epcfg EndpointConfig) *Endpoint {
	chain, err := ParseChain(epcfg.Chain)
	if err != nil {
		panic(err)
	}
	ep := &Endpoint{
		Config:    epcfg,
		Name:      name,
		Chain:     chain,
		Unhealthy: false,
		connected: true}

	if epcfg.SkipMethods != nil {
		ep.SkipMethods = make(map[string]bool)
		for _, meth := range epcfg.SkipMethods {
			ep.SkipMethods[meth] = true
		}
	}
	return ep
}

func (self Endpoint) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain": self.Chain.String(),
		"name":  self.Name,
	})
}

func (self Endpoint) prometheusLabels() prometheus.Labels {
	return prometheus.Labels{
		"chain":    self.Chain.String(),
		"endpoint": self.Name,
	}
}

func (self *Endpoint) Connect() {
	if self.client == nil {
		tr := &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		}
		self.client = &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		}
	}
}

func (self Endpoint) FullUrl(path string) string {
	if path == "" {
		return self.Config.Url
	} else if strings.HasSuffix(self.Config.Url, "/") {
		return self.Config.Url + path[1:]
	} else {
		return self.Config.Url + path
	}
}

// RESTful methods
func (self *Endpoint) PipeRequest(rootCtx context.Context, path string, w http.ResponseWriter, r *http.Request) error {
	self.Connect()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// prepare request
	// TODO: join the server url and method
	url := self.FullUrl(path)
	req, err := http.NewRequestWithContext(ctx, r.Method, url, r.Body)
	if err != nil {
		return errors.Wrap(err, "http.NewRequestWithContext")
	}

	// copy request headers
	for k, vlist := range r.Header {
		for _, v := range vlist {
			req.Header.Add(k, v)
		}
	}

	if self.Config.Headers != nil {
		for k, v := range self.Config.Headers {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	resp, err := self.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "http Do")
	}

	// pipe the response
	for hn, hvs := range resp.Header {
		for _, hv := range hvs {
			w.Header().Add(hn, hv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return nil
}

// Perform a GET request and process the response as JSON
func (self *Endpoint) GetJson(rootCtx context.Context, path string, headers map[string]string, output interface{}) error {
	return self.RequestJson(rootCtx, "GET", path, nil, headers, output)
}

// encode types of body to bytes
// case body is []byte then return it intactly
// case body is struct then return JSON marshalling
func (self Endpoint) encodeBody(body interface{}) ([]byte, string, error) {
	if body == nil {
		return nil, "", nil
	} else if data, ok := body.([]byte); ok {
		return data, "", nil
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		return data, "application/json", nil
	}
	// TODO: handle http form
}

// Post body and parse JSON result, the request body can be in form of
// bytes, map and golang struct
func (self *Endpoint) PostJson(rootCtx context.Context, path string, body interface{}, headers map[string]string, output interface{}) error {
	data, ctype, err := self.encodeBody(body)
	if err != nil {
		return errors.Wrap(err, "encodeBody")
	}
	if ctype != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = ctype
	}
	return self.RequestJson(rootCtx, "POST", path, data, headers, output)
}

// Generic way of performing a HTTP request and shift the response as JSON
func (self *Endpoint) RequestJson(rootCtx context.Context, method string, path string, data []byte, headers map[string]string, output interface{}) error {
	self.Connect()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// prepare request
	// TODO: join the server url and path
	url := self.FullUrl(path)
	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return errors.Wrap(err, "http.NewRequestWithContext")
	}

	req.Header.Set("Accept", "application/json")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := self.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "get Do")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		self.Log().Warnf("invalid response status %d", resp.StatusCode)
		abnResp := &jsonzhttp.WrappedResponse{
			Response: resp,
		}
		return errors.Wrap(abnResp, "abnormal response")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll")
	}

	err = json.Unmarshal(respBody, output)
	if err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}
	return nil
}

// graphql
type gqlRequest struct {
	Query     string
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   *json.RawMessage `json:"data,omitempty"`
	Errors interface{}      `json:"errors,omitempty"`
}

type GqlErrors struct {
	Errors interface{}
}

func (self GqlErrors) Error() string {
	return fmt.Sprintf("%#v\n", self.Errors)
}

func (self *Endpoint) RequestGraphQL(ctx context.Context, query string, variables map[string]interface{}, headers map[string]string, output interface{}) error {
	req := gqlRequest{
		Query:     query,
		Variables: variables,
	}
	var resp gqlResponse
	err := self.PostJson(ctx, "", req, headers, &resp)
	if err != nil {
		return err
	}
	if resp.Errors != nil {
		return &GqlErrors{Errors: resp.Errors}
	} else if resp.Data != nil {
		err := json.Unmarshal(*resp.Data, output)
		if err != nil {
			return err
		}
	}
	return nil

}

func (self *Endpoint) GetClientVersion(ctx context.Context) {
	delegator := GetDelegatorFactory().GetChaintipDelegator(self.Chain.Brand)
	version, err := delegator.GetClientVersion(ctx, self)
	if err != nil {
		self.Log().Warnf("error while getting client version %s", err)
	} else if version != "" {
		self.Log().Infof("client version set to %s", version)
		self.ClientVersion = version
	}
}

func (self Endpoint) Info() EndpointInfo {
	return EndpointInfo{
		Name:      self.Name,
		Chain:     self.Chain.String(),
		Unhealthy: self.Unhealthy,
		Chaintip:  self.Chaintip,
	}
}
