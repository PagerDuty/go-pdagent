package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"go.uber.org/zap"
)

const url = "https://api.pagerduty.com/agent/2014-03-14/heartbeat/go-pdagent"
const frequencySeconds = 60 * 60 // Send heartbeat every hour

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

type HeartbeatResponseBody struct {
	HeartBeatIntervalSeconds int `json:"heartbeat_interval_secs"`
}

func NewHeartbeat() Heartbeat {
	hb := heartbeat{
		ticker:   nil,
		shutdown: make(chan bool),
		logger:   common.Logger.Named("Heartbeat"),
		client: &http.Client{
			Transport: common.NewRetryTransport(),
			Timeout:   1 * time.Minute,
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

	statusCode, err := hb.makeHeartbeatRequest()
	if statusCode/100 == 2 {
		hb.logger.Info("Heartbeat sent successful")
	} else {
		hb.logger.Warnf("Heartbeat request returned a non-success response code: %s", statusCode)
	}

	if err != nil {
		hb.logger.Warnf("An error occurred while sending heartbeat: ", err)
	}
}

func (hb *heartbeat) makeHeartbeatRequest() (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("User-Agent", common.UserAgent())
	req.Header.Add("Accept", "application/json")

	httpResp, err := hb.client.Do(req)
	if err != nil {
		return 0, err
	}

	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return 0, err
	}

	var responseBody HeartbeatResponseBody

	err = json.Unmarshal(respBody, &responseBody)
	if err != nil {
		return 0, err
	}

	hb.logger.Info("Updating heartbeat frequency to ", responseBody.HeartBeatIntervalSeconds)
	hb.ticker.Stop()
	hb.ticker = time.NewTicker(time.Duration(responseBody.HeartBeatIntervalSeconds) * time.Second)

	return httpResp.StatusCode, nil
}
