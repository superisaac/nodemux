package balancer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func (self AbnormalResponse) Error() string {
	return fmt.Sprintf("Abnormal response %d", self.Response.StatusCode)
}

// EndpointSet
func (self *EndpointSet) ResetMaxTipHeight() {
	maxHeight := 0
	for _, epItem := range self.items {
		if epItem.Tip != nil && epItem.Tip.Height > maxHeight {
			maxHeight = epItem.Tip.Height
		}
	}
	self.maxTipHeight = maxHeight
}

func (self EndpointSet) prometheusLabels(chain string, network string) prometheus.Labels {
	return prometheus.Labels{
		"chain":   chain,
		"network": network,
	}
}

/// Create an empty endpoint
func NewEndpoint() *Endpoint {
	return &Endpoint{Healthy: true}
}

func (self Endpoint) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain":   self.Chain.Name,
		"network": self.Chain.Network,
		"name":    self.Name,
	})
}

func (self Endpoint) prometheusLabels() prometheus.Labels {
	return prometheus.Labels{
		"chain":    self.Chain.Name,
		"network":  self.Chain.Network,
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

func (self *Endpoint) CallRPC(rootCtx context.Context, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
	self.Connect()

	traceId := reqmsg.TraceId()

	reqmsg.SetTraceId("")

	marshaled, err := jsonrpc.MessageBytes(reqmsg)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(marshaled)

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", self.ServerUrl, reader)
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequestWithContext")
	}
	//req.Header.Add("X-Trace-Id", traceId)
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")
	if self.Headers != nil {
		for k, v := range self.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http Do")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		self.Log().Warnf("invalid response status %d", resp.StatusCode)
		abnResp := &AbnormalResponse{
			//Code: resp.StatusCode,
			// TODO: filter scam headers
			//Header: resp.Header,
			//Body:   body,
			Response: resp,
		}
		return nil, errors.Wrap(abnResp, "abnormal response")
		//return nil, errors.New(fmt.Sprintf("bad resp %d", resp.StatusCode))
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll")
	}

	respMsg, err := jsonrpc.ParseBytes(respBody)
	if err != nil {
		return nil, err
	}
	respMsg.SetTraceId(traceId)
	return respMsg, nil
} // CallHTTP

func (self Endpoint) FullUrl(path string) string {
	if strings.HasSuffix(self.ServerUrl, "/") {
		return self.ServerUrl + path[1:]
	} else {
		return self.ServerUrl + path
	}
}

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

	if self.Headers != nil {
		for k, v := range self.Headers {
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

func (self *Endpoint) GetJson(rootCtx context.Context, path string, headers map[string]string, target interface{}) error {
	return self.RequestJson(rootCtx, "GET", path, nil, headers, target)
}

func (self *Endpoint) RequestJson(rootCtx context.Context, method string, path string, data []byte, headers map[string]string, target interface{}) error {
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
	//req.Header.Set("X-Forward-For", r.RemoteAddr)
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
		abnResp := &AbnormalResponse{
			Response: resp,
		}
		return errors.Wrap(abnResp, "abnormal response")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadAll")
	}

	err = json.Unmarshal(respBody, target)
	if err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}
	return nil
}
