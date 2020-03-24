package server

import (
	"encoding/json"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"io/ioutil"
	"net/http"
)

func (s *Server) SendHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorResp(rw, 400, []string{err.Error()})
		return
	}

	s.logger.Debugf("/send payload: %v", string(body))

	var event eventsapi.EventV2
	if err = json.Unmarshal(body, &event); err != nil {
		errorResp(rw, 400, []string{err.Error()})
		return
	}

	key, err := s.Queue.Enqueue(&event)
	if err != nil {
		errorResp(rw, 500, []string{err.Error()})
		return
	}

	okResp(rw, SendResponse{Key: key})
}

type SendResponse struct {
	Key string `json:"key"`
}
