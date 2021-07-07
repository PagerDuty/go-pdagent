package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
)

func (s *Server) SendHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errorResp(rw, 400, []string{err.Error()})
		return
	}

	s.logger.Debugf("/send payload: %v", string(body))

	eventContainer := eventsapi.EventContainer{
		EventVersion: eventsapi.StringToEventVersion[req.Header["Pd-Event-Version"][0]],
	}

	if err = json.Unmarshal(body, &eventContainer.EventData); err != nil {
		errorResp(rw, 400, []string{err.Error()})
		return
	}

	key, err := s.Queue.Enqueue(&eventContainer)
	if err != nil {
		errorResp(rw, 500, []string{err.Error()})
		return
	}

	okResp(rw, SendResponse{Key: key})
}

type SendResponse struct {
	Key string `json:"key"`
}
