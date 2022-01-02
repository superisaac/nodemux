package balancer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/jsonrpc"
	"io/ioutil"
	"net/http"
	"time"
)

func (self AbnormalResponse) Error() string {
	return fmt.Sprintf("Abnormal response %d %s %s", self.Code, self.ContentType, self.Body[0:30])
}

/// Create an empty endpoint
func NewEndpoint() *Endpoint {
	ep := &Endpoint{Healthy: true, HeightPadding: 2}
	return ep
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

func (self *Endpoint) CallHTTP(rootCtx context.Context, reqmsg *jsonrpc.RequestMessage) (jsonrpc.IMessage, error) {
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
	req.Header.Add("X-Trace-Id", traceId)

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http Do")
	}
	if resp.StatusCode != 200 {
		self.Log().Warnf("invalid response status %d", resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "ioutil.ReadAll")
		}
		errResp := &AbnormalResponse{
			Code:        resp.StatusCode,
			ContentType: resp.Header.Get("Content-Type"),
			Body:        body,
		}
		return nil, errors.Wrap(errResp, "abnormal response")
		//return nil, errors.New(fmt.Sprintf("bad resp %d", resp.StatusCode))
	}
	defer resp.Body.Close()
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
}

func (self Endpoint) Log() *log.Entry {
	return log.WithFields(log.Fields{
		"chain":   self.Chain.Name,
		"network": self.Chain.Network,
		"name":    self.Name,
	})
}
