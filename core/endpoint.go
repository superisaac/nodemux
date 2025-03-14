package nodemuxcore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsoff/net"
)

// / Create an endpoint instance
func NewEndpoint(name string, epcfg EndpointConfig) *Endpoint {
	chain, err := ParseChain(epcfg.Chain)
	if err != nil {
		panic(err)
	}
	hash := sha256.New()
	hash.Write([]byte(epcfg.Url))
	checksum := hash.Sum(nil)
	urlDigest := hex.EncodeToString(checksum)

	ep := &Endpoint{
		Config:    epcfg,
		Name:      name,
		URLDigest: urlDigest,
		Chain:     chain,
		Healthy:   true,
		connected: true}

	if epcfg.SkipMethods != nil {
		ep.SkipMethods = make(map[string]bool)
		for _, meth := range epcfg.SkipMethods {
			ep.SkipMethods[meth] = true
		}
	}
	return ep
}

func (ep Endpoint) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain":    ep.Chain.String(),
		"endpoint": ep.Name,
	})
}

func (ep Endpoint) prometheusLabels() prometheus.Labels {
	return prometheus.Labels{
		"chain":    ep.Chain.String(),
		"endpoint": ep.Name,
	}
}

func (ep Endpoint) incrRelayCount() {
	metricsEndpointRelayCount.With(prometheus.Labels{
		"chain":    ep.Chain.String(),
		"endpoint": ep.Name,
	}).Inc()
}

func (ep Endpoint) incrBlockheadCount() {
	metricsBlockheadCount.With(prometheus.Labels{
		"chain":    ep.Chain.String(),
		"endpoint": ep.Name,
	}).Inc()
}

func (ep *Endpoint) Connect() {
	if ep.client == nil {
		tr := &http.Transport{
			MaxIdleConns:        30,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		}
		timeout := ep.Config.Timeout
		if timeout <= 0 {
			timeout = 90
		}
		ep.client = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(timeout) * time.Second,
		}
	}
}

func (ep Endpoint) FullUrl(path string) string {
	if path == "" {
		return ep.Config.Url
	} else if strings.HasSuffix(ep.Config.Url, "/") && strings.HasPrefix(path, "/") {
		return ep.Config.Url + path[1:]
	} else {
		return ep.Config.Url + path
	}
}

// RESTful methods
func (ep *Endpoint) PipeRequest(rootCtx context.Context, path string, w http.ResponseWriter, r *http.Request) error {
	resp, err := ep.doResponse(rootCtx, path, r)
	if err != nil {
		if os.IsTimeout(err) {
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("timeout"))
			return nil
		}
		return errors.Wrap(err, "http Do")
	}

	// pipe the response
	for hn, hvs := range resp.Header {
		if strings.ToLower(hn) == "server" {
			w.Header().Set("Server", "nodemux")
		} else {
			for _, hv := range hvs {
				w.Header().Set(hn, hv)
			}
		}
	}
	w.Header().Set("X-Real-Endpoint", ep.Name)
	w.WriteHeader(resp.StatusCode)
	if written, err := io.Copy(w, resp.Body); err != nil {
		ep.Log().WithFields(log.Fields{
			"written": written,
			"path":    path,
		}).Warnf("io copy error %#v", err)
		return err
	}

	return nil
}

func (ep *Endpoint) doResponse(rootCtx context.Context, path string, r *http.Request) (*http.Response, error) {
	ep.Connect()
	ep.incrRelayCount()
	// prepare request

	url := ep.FullUrl(path)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(rootCtx, r.Method, url, io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequestWithContext")
	}

	// copy request headers
	for k, vlist := range r.Header {
		for _, v := range vlist {
			req.Header.Add(k, v)
		}
	}

	if ep.Config.Headers != nil {
		for k, v := range ep.Config.Headers {
			req.Header.Set(k, v)
		}
	}
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	start := time.Now()
	resp, err := ep.client.Do(req)
	delta := time.Since(start)
	fields := log.Fields{
		"method":      path,
		"httpMethod":  r.Method,
		"timeSpentMS": delta.Milliseconds(),
	}
	if err != nil {
		fields["err"] = err.Error()
	}

	if resp != nil {
		fields["status"] = resp.StatusCode
	}

	ep.Log().WithFields(fields).Info("relay http")
	return resp, err

}

// Perform a GET request and process the response as JSON
func (ep *Endpoint) GetJson(rootCtx context.Context, path string, headers map[string]string, output interface{}) error {
	return ep.RequestJson(rootCtx, "GET", path, nil, headers, output)
}

// encode types of body to bytes
// case body is []byte then return it intactly
// case body is struct then return JSON marshalling
func (ep Endpoint) encodeBody(body interface{}) ([]byte, string, error) {
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
func (ep *Endpoint) PostJson(rootCtx context.Context, path string, body interface{}, headers map[string]string, output interface{}) error {
	data, ctype, err := ep.encodeBody(body)
	if err != nil {
		return errors.Wrap(err, "encodeBody")
	}
	if ctype != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = ctype
	}
	return ep.RequestJson(rootCtx, "POST", path, data, headers, output)
}

// Generic way of performing a HTTP request and shift the response as JSON
func (ep *Endpoint) RequestJson(rootCtx context.Context, method string, path string, data []byte, headers map[string]string, output interface{}) error {
	ep.Connect()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// prepare request
	// TODO: join the server url and path
	url := ep.FullUrl(path)
	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return errors.Wrap(err, "http.NewRequestWithContext")
	}

	req.Header.Set("Accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := ep.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "get Do")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		ep.Log().Warnf("invalid response status %d", resp.StatusCode)
		abnResp := &jsoffnet.WrappedResponse{
			Response: resp,
		}
		return errors.Wrap(abnResp, "abnormal response")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "io.ReadAll")
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

func (err GqlErrors) Error() string {
	return fmt.Sprintf("%#v\n", err.Errors)
}

func (ep *Endpoint) RequestGraphQL(ctx context.Context, query string, variables map[string]interface{}, headers map[string]string, output interface{}) error {
	req := gqlRequest{
		Query:     query,
		Variables: variables,
	}
	var resp gqlResponse
	err := ep.PostJson(ctx, "", req, headers, &resp)
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

func (ep *Endpoint) GetClientVersion(ctx context.Context) {
	delegator := GetDelegatorFactory().GetBlockheadDelegator(ep.Chain.Namespace)
	version, err := delegator.GetClientVersion(ctx, ep)
	if err != nil {
		ep.Log().Warnf("error while getting client version %s", err)
	} else if version != "" {
		ep.Log().Infof("client version set to %s", version)
		ep.ClientVersion = version
	}
}

func (ep Endpoint) Info() EndpointInfo {
	return EndpointInfo{
		Name:          ep.Name,
		URLDigest:     ep.URLDigest,
		Chain:         ep.Chain.String(),
		Healthy:       ep.Healthy,
		Blockhead:     ep.Blockhead,
		ClientVersion: ep.ClientVersion,
	}
}

func (ep Endpoint) Available(method string, minHeight int) bool {
	if !ep.Healthy {
		return false
	}

	if minHeight > 0 {
		if ep.Blockhead == nil || ep.Blockhead.Height < minHeight {
			return false
		}
	}

	if method != "" && ep.SkipMethods != nil {
		if _, ok := ep.SkipMethods[method]; ok {
			// the method is not provided by the endpoint, so skip it
			return false
		}
	}
	return true
}
