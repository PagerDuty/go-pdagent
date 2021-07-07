package eventqueue

import (
	"sync"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"go.uber.org/zap"
)

const DefaultBufferSize = 1000

// EventQueues are a basic thread-safe queue for processing PagerDuty events.
//
// Each EventQueue is internally composed of several individual queues
// segmented by routing key, ensuring that events are in-order on a per
// routing key basis. Each of these queues has a single dedicated worker.
//
// All responses occur through a single user-provided channel when enqueuing
// events.
//
// EventQueues also have a configurable, synchronous processor. By default this
// processor sends events to PagerDuty's events API..
//
// Example usage:
//
//     queue := eventqueue.NewEventQueue()
//
//     event := eventsapi.EventV2{
//		   RoutingKey:  "NN2EZIJQPVVQF3KIKN2MDMSUV7E6GLFN",
//		   EventAction: "trigger",
//		   Payload:     eventsapi.PayloadV2{
//			   Summary:  "Test summary",
//			   Source:   "Test source",
//			   Severity: "Error",
//		   },
//     }
//
//     respChan = make(chan eventqueue.Response)
//
//     queue.Enqueue(event, respChan)
//
//     resp := <-respChan
//
//     // When you're done with the queue.
//     queue.Shutdown()
type EventQueue struct {
	Processor Processor

	logger *zap.SugaredLogger
	mu     sync.Mutex
	queues map[string]chan Job
	stop   chan bool
	wg     sync.WaitGroup
}

// NewEventQueue initializes a new default EventQueue.
func NewEventQueue() *EventQueue {
	logger := common.Logger.Named("EventQueue")
	logger.Info("Creating new EventQueue.")

	return &EventQueue{
		Processor: DefaultProcessor,
		logger:    logger,
		queues:    make(map[string]chan Job),
		stop:      make(chan bool),
	}
}

// Shutdown the queue and all associated workers.
//
// There may be a blocking delay while any running workers or processors
// attempt to complete their current tasks.
func (q *EventQueue) Shutdown() {
	q.logger.Info("Shutting down EventQueue.")
	for _, w := range q.queues {
		close(w)
	}
	q.wg.Wait()
	close(q.stop)
	q.logger.Info("Shut down EventQueue.")
}

// Enqueue a PagerDuty event for processing.
//
// Accepts an event and a channel over which to communicate responses. Errors
// come in two flavors: Synchronous errors (e.g. event is invalid and never
// queued) as a return value and asynchronous errors (e.g. server error) that
// are part of the channel Response.
func (q *EventQueue) Enqueue(eventContainer *eventsapi.EventContainer, respChan chan<- Response) error {
	if err := eventContainer.GetEvent().Validate(); err != nil {
		return err
	}

	key := eventContainer.GetEvent().GetRoutingKey()

	q.ensureWorker(key)

	select {
	case q.queues[key] <- Job{eventContainer, respChan, q.logger.Named(key)}:
		return nil
	default:
		respChan <- Response{Error: &ErrBufferOverflow{key, DefaultBufferSize}}
		return nil
	}
}

func (q *EventQueue) ensureWorker(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queues[key] != nil {
		return
	}

	c := make(chan Job, DefaultBufferSize)
	q.wg.Add(1)
	go q.worker(key, c)
	q.queues[key] = c
}

func (q *EventQueue) worker(key string, c <-chan Job) {
	defer q.wg.Done()
	logger := q.logger.Named(key)

	logger.Infof("Worker started.")
	for job := range c {
		logger.Infof("Job started, %v pending.", len(c))
		q.Processor(job, q.stop)
	}
	logger.Infof("Worker stopped.")
}

type Job struct {
	EventContainer *eventsapi.EventContainer
	ResponseChan   chan<- Response
	Logger         *zap.SugaredLogger
}

type Response struct {
	Response eventsapi.Response
	Error    error
}
