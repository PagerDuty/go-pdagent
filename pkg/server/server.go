package server

import (
	"context"
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"github.com/PagerDuty/pagerduty-agent/pkg/persistentqueue"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Queue interface {
	Enqueue(eventsapi.Event) (string, error)
	Retry(string) (int, error)
	Shutdown() error
	Start() error
	Status(string) ([]persistentqueue.StatusItem, error)
}

type Server struct {
	HTTPServer *http.Server
	Queue      Queue

	secret string
	logger *zap.SugaredLogger
}

type Option func(*Server)

func NewServer(address, secret string, queue Queue) *Server {
	logger := common.Logger.Named("Server")

	server := Server{
		HTTPServer: &http.Server{
			Addr:           address,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Queue:  queue,
		secret: secret,
		logger: logger,
	}

	server.HTTPServer.Handler = Router(&server)

	return &server
}

func (s *Server) Start() error {
	s.logger.Infof("Server starting at %v", s.HTTPServer.Addr)

	if err := s.Queue.Start(); err != nil {
		s.logger.Error("Failed to start server's queue.")
		return err
	}

	go func() {
		s.logger.Info(s.HTTPServer.ListenAndServe())
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		s.logger.Error(err)
	}

	if err := s.Queue.Shutdown(); err != nil {
		s.logger.Error("Error shutting down server's queue.")
		return err
	}

	os.Exit(0)
	return nil
}
