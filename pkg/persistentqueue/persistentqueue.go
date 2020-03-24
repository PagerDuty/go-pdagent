package persistentqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventqueue"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"github.com/asdine/storm"
	"go.uber.org/zap"
	"sync"
)

type EventQueue interface {
	Shutdown()
	Enqueue(eventsapi.Event, chan<- eventqueue.Response) error
}

type PersistentQueue struct {
	DB         *storm.DB
	Events     storm.Node
	EventQueue EventQueue

	logger *zap.SugaredLogger
	wg     sync.WaitGroup
}

type Option func(*PersistentQueue)

func WithEventQueue(eq EventQueue) Option {
	return func(pq *PersistentQueue) {
		pq.EventQueue = eq
	}
}

func NewPersistentQueue(path string, options ...Option) (*PersistentQueue, error) {
	logger := common.Logger.Named("PersistentQueue")
	logger.Info("Creating new PersistentQueue.")

	q := PersistentQueue{
		logger: logger,
	}

	for _, option := range options {
		option(&q)
	}

	if q.EventQueue == nil {
		q.EventQueue = eventqueue.NewEventQueue()
	}

	err := q.Start(path)
	if err != nil {
		logger.Error("Error starting PersistentQueue: ", err)
		return nil, err
	}

	return &q, nil
}

func (q *PersistentQueue) Start(path string) error {
	db, err := storm.Open(path)
	if err != nil {
		return err
	}

	q.DB = db
	q.Events = q.DB.From("events")

	var pendingEvents []Event
	if err := q.Events.Find("Status", StatusPending, &pendingEvents); err != nil && err != storm.ErrNotFound {
		q.logger.Error("Error querying for pending events: ", err)
		return err
	}

	q.logger.Infof("Enqueuing %v pending events.", len(pendingEvents))
	for _, e := range pendingEvents {
		q.processEvent(&e)
	}

	return nil
}

// Stop a `PersistentQueue`, performing any necessary cleanup.
func (q *PersistentQueue) Shutdown() error {
	q.logger.Info("Shutting down PersistentQueue.")
	q.EventQueue.Shutdown()
	q.wg.Wait()
	if err := q.DB.Close(); err != nil {
		return err
	}
	q.logger.Info("Shut down PersistentQueue.")
	return nil
}
