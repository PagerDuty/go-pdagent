package server

import (
	"github.com/gorilla/mux"
)

func Router(s *Server) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", s.HealthHandler)
	r.HandleFunc("/send", s.SendHandler)
	r.HandleFunc("/queue/retry", s.RetryHandler)
	r.HandleFunc("/queue/status", s.StatusHandler)

	r.Use(loggingMiddleware(s.logger))
	r.Use(authMiddleware(s))

	return r
}
