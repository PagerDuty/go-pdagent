package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/PagerDuty/go-pdagent/pkg/persistentqueue"
	"go.uber.org/zap"
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

	pidfile string
	secret  string
	logger  *zap.SugaredLogger
}

type Option func(*Server)

func NewServer(address, secret, pidfile string, queue Queue) *Server {
	logger := common.Logger.Named("Server")

	server := Server{
		HTTPServer: &http.Server{
			Addr:           address,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Queue:   queue,
		pidfile: pidfile,
		secret:  secret,
		logger:  logger,
	}

	server.HTTPServer.Handler = Router(&server)

	return &server
}

func (s *Server) Start() error {
	s.logger.Infof("Server starting at %v", s.HTTPServer.Addr)

	if err := s.initPidfile(); err != nil {
		return err
	}

	if err := s.Queue.Start(); err != nil {
		s.logger.Error("Failed to start server's queue.")
		return err
	}

	go func() {
		s.logger.Info(s.HTTPServer.ListenAndServe())
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
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

	if err := common.RemovePidfile(s.pidfile); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func (s *Server) initPidfile() error {
	if err := os.MkdirAll(path.Dir(s.pidfile), 0744); err != nil {
		return err
	}

	if err := common.InitPidfile(s.pidfile); err == common.ErrPidfileExists {
		s.logger.Errorf("Pidfile already exists, suggesting server is already running: %v", s.pidfile)
		return err
	} else if err != nil {
		s.logger.Errorf("Error encountered writing pidfile: %v", err)
		return err
	}

	s.logger.Infof("Successfully wrote pidfile: %v", s.pidfile)
	return nil
}
