package server

import (
	"context"
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/persistentqueue"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	HTTPServer *http.Server
	Queue      *persistentqueue.PersistentQueue

	database string
	secret   string
	logger   *zap.SugaredLogger
}

type Option func(*Server)

func NewServer(address, secret, database string) *Server {
	logger := common.Logger.Named("Server")

	server := Server{
		database: database,
		secret:   secret,
		logger:   logger,
	}

	handler := Router(&server)

	server.HTTPServer = &http.Server{
		Addr:           address,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &server
}

func (s *Server) Start() error {
	s.logger.Infof("Server starting at %v", s.HTTPServer.Addr)

	go func() {
		s.logger.Info(s.HTTPServer.ListenAndServe())
	}()

	queue, err := persistentqueue.NewPersistentQueue(s.database)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	s.Queue = queue

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		s.logger.Error(err)
	}

	queue.Shutdown()

	os.Exit(0)
	return nil
}
