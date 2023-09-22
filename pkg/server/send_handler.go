package server

import (
	"io"
	"net/http"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
)

func (s *Server) SendHandler(rw http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		errorResp(rw, 400, []string{err.Error()})
		return
	}

	s.logger.Debugf("/send payload: %v", string(body))

	eventContainer := eventsapi.EventContainer{
		EventVersion: eventsapi.StringToEventVersion[req.Header["Pd-Event-Version"][0]],
		EventData:    body,
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
