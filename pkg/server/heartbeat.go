package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"go.uber.org/zap"
)

const url = "https://api.pagerduty.com/agent/2014-03-14/heartbeat/go-pdagent"
const frequencySeconds = 60 * 60 // Send heartbeat every hour
const maxRetries = 3
const maxRetryInterval = 15 * time.Second

var ErrHeartbeatError = errors.New("an error was encountered while sending the heartbeat")

type Heartbeat interface {
	Start()
	Shutdown()
}

type heartbeat struct {
	ticker    *time.Ticker
	shutdown  chan bool
	logger    *zap.SugaredLogger
	client    *http.Client
	frequency int
}

type heartbeatResponseBody struct {
	HeartBeatIntervalSeconds int `json:"heartbeat_interval_secs"`
}

func NewHeartbeat() Heartbeat {
	transport := common.NewRetryTransport()
	transport.MaxRetries = maxRetries
	transport.MaxInterval = maxRetryInterval

	hb := heartbeat{
		ticker:   nil,
		shutdown: make(chan bool),
		logger:   common.Logger.Named("Heartbeat"),
		client: &http.Client{
			Transport: transport,
			Timeout:   20 * time.Second,
		},
		frequency: frequencySeconds,
	}

	return &hb
}

func (hb *heartbeat) Start() {
	hb.logger.Info("Starting heartbeat.")
	hb.ticker = time.NewTicker(time.Duration(hb.frequency) * time.Second)

	go func() {
		for {
			select {
			case <-hb.shutdown:
				return
			case <-hb.ticker.C:
				go hb.beat()
			}
		}
	}()
}

func (hb *heartbeat) Shutdown() {
	hb.ticker.Stop()
	hb.shutdown <- true
	hb.logger.Info("Heartbeat stopped.")
}

func (hb *heartbeat) beat() {
	hb.logger.Info("Sending heartbeat")

	heartbeatResponse, err := hb.doHeartbeatRequest()
	if err != nil {
		hb.logger.Warnf("An error occurred while sending heartbeat: ", err)
		return
	}

	hb.logger.Info("Heartbeat successful")

	hb.logger.Info("Updating heartbeat frequency to ", heartbeatResponse.HeartBeatIntervalSeconds)
	hb.ticker.Stop()
	hb.ticker = time.NewTicker(time.Duration(heartbeatResponse.HeartBeatIntervalSeconds) * time.Second)
}

func (hb *heartbeat) doHeartbeatRequest() (*heartbeatResponseBody, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", common.UserAgent())
	req.Header.Add("Accept", "application/json")

	httpResp, err := hb.client.Do(req)
	if !common.IsSuccessResponse(httpResp, err) {
		return nil, ErrHeartbeatError
	}

	defer httpResp.Body.Close()
	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	var responseBody heartbeatResponseBody
	err = json.Unmarshal(respBody, &responseBody)
	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}
