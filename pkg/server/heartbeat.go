package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"go.uber.org/zap"
)

const HEARTBEAT_URL = "https://api.pagerduty.com/agent/2014-03-14/heartbeat/go-pdagent"
const HEARTBEAT_FREQUENCY_SECONDS = 60 * 60 // Send heartbeat every hour
const HEARTBEAT_MAX_RETRIES = 10
const RETRY_GAP_SECONDS = 10

type HeartbeatTask struct {
	ticker      *time.Ticker
	shutdown    chan bool
	logger      *zap.SugaredLogger
	client      *http.Client
	agentIdFile string
}

func NewHeartbeatTask() *HeartbeatTask {
	hb := HeartbeatTask{
		ticker:      nil,
		shutdown:    make(chan bool),
		logger:      common.Logger.Named("Heartbeat"),
		client:      &http.Client{},
		agentIdFile: "",
	}

	return &hb
}

func (hb *HeartbeatTask) Start(agentIdFile string) {
	hb.logger.Info("Starting heartbeat")
	hb.ticker = time.NewTicker(HEARTBEAT_FREQUENCY_SECONDS * time.Second)
	hb.agentIdFile = agentIdFile

	go func() {
		for {
			select {
			case <-hb.shutdown:
				hb.logger.Info("Heartbeat goroutine shutdown")
				return
			case <-hb.ticker.C:
				go hb.beat()
			}
		}
	}()
}

func (hb *HeartbeatTask) Shutdown() {
	hb.ticker.Stop()
	hb.shutdown <- true
	hb.logger.Info("Heartbeat stopped")
}

func (hb *HeartbeatTask) beat() {
	hb.logger.Info("Sending heartbeat")

	attempts := 0
	maxRetryTime := time.Now().Add(time.Second * HEARTBEAT_FREQUENCY_SECONDS / 2)

	for {
		attempts++

		req, err := http.NewRequest("GET", HEARTBEAT_URL, nil)
		if err != nil {
			hb.logger.Error("Failed to create heartbeat request - will not retry")
			return
		}

		req.Header.Add("User-Agent", userAgent(*hb))
		req.Header.Add("Content-Type", "application/json")

		httpResp, err := hb.client.Do(req)
		if err != nil {
			hb.logger.Error("Failed to send heartbeat request - will not retry")
			return
		}

		httpResp.Body.Close()

		if httpResp.StatusCode/100 == 2 {
			hb.logger.Info("Heartbeat successful!")
			return
		} else if httpResp.StatusCode/100 == 5 {
			hb.logger.Error("Error sending heartbeat - will retry")
		} else {
			hb.logger.Info("Heartbeat request returned a non-success response code - will retry")
		}

		if attempts >= HEARTBEAT_MAX_RETRIES {
			hb.logger.Info("Heartbeat retry limit exceeded - will not retry")
			return
		}

		if time.Now().After(maxRetryTime) {
			hb.logger.Info("Heartbeat retry time limit exceeded - will not retry")
			return
		}

		hb.logger.Info("Sleeping before retry")
		time.Sleep(RETRY_GAP_SECONDS * time.Second)
		hb.logger.Info("Retrying heartbeat")
	}
}

func userAgent(heartbeat HeartbeatTask) string {
	version := common.Version
	agentId := common.GetAgentId(heartbeat.agentIdFile)

	return fmt.Sprintf("go-pdagent/%v (Agent ID: %s)", version, agentId)
}
