package server

import (
	"fmt"
	"net/http"
)

func (s *Server) RetryHandler(rw http.ResponseWriter, req *http.Request) {
	rk := req.URL.Query().Get("rk")

	if rk == "" {
		s.logger.Debugf("Retrying for all routing keys.")
	} else {
		s.logger.Debugf("Retrying for routing key %v", rk)
	}

	count, err := s.Queue.Retry(rk)
	if err != nil {
		errorResp(rw, 500, []string{err.Error()})
		return
	}

	okResp(rw, RetryResponse{fmt.Sprintf("Retrying %v events.", count)})
}

type RetryResponse struct {
	Message string `json:"message"`
}
