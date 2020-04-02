package persistentqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventqueue"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"github.com/asdine/storm"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"sync"
)

type EventQueue interface {
	Enqueue(eventsapi.Event, chan<- eventqueue.Response) error
	Shutdown()
}

type PersistentQueue struct {
	DB         *storm.DB
	Events     storm.Node
	EventQueue EventQueue

	path   string
	logger *zap.SugaredLogger
	tmp    bool
	wg     sync.WaitGroup
}

type Option func(*PersistentQueue)

func WithFile(path string) Option {
	return func(q *PersistentQueue) {
		q.path = path
		q.tmp = false
	}
}

func WithEventQueue(eq EventQueue) Option {
	return func(q *PersistentQueue) {
		q.EventQueue = eq
	}
}

func NewPersistentQueue(options ...Option) *PersistentQueue {
	logger := common.Logger.Named("PersistentQueue")
	logger.Info("Creating new PersistentQueue.")

	q := PersistentQueue{
		EventQueue: eventqueue.NewEventQueue(),
		logger:     logger,
		tmp:        true,
	}

	for _, option := range options {
		option(&q)
	}

	return &q
}

func (q *PersistentQueue) Start() error {
	if q.tmp {
		dbFile, err := ioutil.TempFile("", "pagerduty-agent.*.db")
		if err != nil {
			return err
		}
		q.path = dbFile.Name()
		dbFile.Close()
	}

	db, err := storm.Open(q.path)
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

	if q.tmp {
		os.Remove(q.path)
	}

	q.logger.Info("Shut down PersistentQueue.")
	return nil
}
