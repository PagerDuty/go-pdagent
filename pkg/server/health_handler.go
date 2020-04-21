package server

import (
	"fmt"
	"net/http"
)

func (s *Server) HealthHandler(rw http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(rw, "OK")
	if err != nil {
		s.logger.Error("Error responding to healthcheck.")
	}
}
