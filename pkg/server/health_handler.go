package server

import "net/http"

func (s *Server) HealthHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Write([]byte("OK"))
}
