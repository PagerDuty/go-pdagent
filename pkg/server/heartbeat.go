package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const url = "https://api.pagerduty.com/agent/2014-03-14/heartbeat/go-pdagent"
const frequencySeconds = 60 * 60 // Send heartbeat every hour
const maxRetries = 10
const retryGapSeconds = 10

type Heartbeat interface {
	Start()
	Shutdown()
}

type heartbeat struct {
	id        string
	ticker    *time.Ticker
	shutdown  chan bool
	logger    *zap.SugaredLogger
	client    *http.Client
	frequency int
}

type HeartbeatResponseBody struct {
	HeartBeatIntervalSeconds int `json:"heartbeat_interval_secs"`
}

func NewHeartbeat() *heartbeat {
	hb := heartbeat{
		id:        uuid.NewString(),
		ticker:    nil,
		shutdown:  make(chan bool),
		logger:    common.Logger.Named("Heartbeat"),
		client:    &http.Client{},
		frequency: frequencySeconds,
	}

	return &hb
}

func (hb *heartbeat) Start() {
	hb.logger.Info("Starting heartbeat")
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
	hb.logger.Info("Heartbeat stopped")
}

func (hb *heartbeat) beat() {
	hb.logger.Info("Sending heartbeat")

	attempts := 0

	for {
		attempts++

		statusCode, isError := hb.makeHeartbeatRequest()
		if isError {
			hb.logger.Error("Failed to send heartbeat request - will not retry")
			return
		}

		if statusCode/100 == 2 {
			hb.logger.Info("Heartbeat successful!")
			return
		} else if statusCode/100 == 5 {
			hb.logger.Error("Error sending heartbeat - will retry")
		} else {
			hb.logger.Info("Heartbeat request returned a non-success response code - will retry")
		}

		if attempts >= maxRetries {
			hb.logger.Info("Heartbeat retry limit exceeded - will not retry")
			return
		}

		hb.logger.Info("Sleeping before retry")
		time.Sleep(retryGapSeconds * time.Second)
		hb.logger.Info("Retrying heartbeat")
	}
}

func (hb *heartbeat) makeHeartbeatRequest() (int, bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, true
	}

	req.Header.Add("User-Agent", userAgent(*hb))
	req.Header.Add("Accept", "application/json")

	httpResp, err := hb.client.Do(req)
	if err != nil {
		return 0, true
	}

	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		hb.logger.Info("Could not read response from heartbeat request.")
	}

	var responseBody HeartbeatResponseBody

	err = json.Unmarshal(respBody, &responseBody)
	if err != nil {
		hb.logger.Info("Could not decode heartbeat response body.")
	} else {
		hb.logger.Info("Updating heartbeat frequency to ", responseBody.HeartBeatIntervalSeconds)
		hb.ticker.Stop()
		hb.ticker = time.NewTicker(time.Duration(responseBody.HeartBeatIntervalSeconds) * time.Second)
	}

	return httpResp.StatusCode, false
}

func userAgent(hb heartbeat) string {
	version := common.Version

	return fmt.Sprintf("go-pdagent/%v (Agent ID: %s)", version, hb.id)
}
