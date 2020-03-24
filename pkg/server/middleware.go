package server

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

func loggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Infof("Handling request: %v", r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}

func authMiddleware(s *Server) func(http.Handler) http.Handler {
	serverHeader := fmt.Sprintf("token %v", s.secret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if s.secret == "" {
				next.ServeHTTP(w, r)
				return
			}

			clientHeader := r.Header.Get("Authorization")
			if clientHeader != serverHeader {
				s.logger.Infof("Authorization failure with client auth header: %v", clientHeader)
				errorResp(w, 401, []string{"Unauthorized, expected matching secret token in Authorization header."})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
