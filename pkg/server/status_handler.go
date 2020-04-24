package server

import (
	"net/http"

	"github.com/PagerDuty/go-pdagent/pkg/persistentqueue"
)

func (s *Server) StatusHandler(rw http.ResponseWriter, req *http.Request) {
	rk := req.URL.Query().Get("rk")

	if rk == "" {
		s.logger.Debugf("Status for all routing keys.")
	} else {
		s.logger.Debugf("Status for routing key %v", rk)
	}

	statusItems, err := s.Queue.Status(rk)
	if err != nil {
		errorResp(rw, 500, []string{err.Error()})
		return
	}

	okResp(rw, StatusResponse{StatusItems: statusItems})
}

type StatusResponse struct {
	StatusItems []persistentqueue.StatusItem `json:"status_items,omitempty"`
}
