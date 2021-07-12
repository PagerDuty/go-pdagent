package persistentqueue

import (
	"github.com/PagerDuty/go-pdagent/pkg/eventqueue"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
)

// Enqueue adds an event to the persistent queue for processing.
//
// Returns the event record's key along with any synchronous errors.
//
// Only synchronous errors (e.g. invalid event) are supported as there are
// cases where we might not have a per-event response channel (e.g. processing
// a backlog).
func (q *PersistentQueue) Enqueue(eventContainer *eventsapi.EventContainer) (string, error) {
	event, err := eventContainer.UnmarshalEvent()
	if err != nil {
		q.logger.Errorf("Failed to unmarshal event container in queue", err)
		return "", err
	}

	if err := event.Validate(); err != nil {
		q.logger.Errorf("Failed to validate event in queue %v.", event.GetRoutingKey(), err)
		return "", err
	}

	e, err := NewEvent(eventContainer)
	if err != nil {
		return "", err
	}
	q.logger.Infof("Enqueuing to %v with key %v.", event.GetRoutingKey(), e.Key)

	if err := e.Create(q.Events); err != nil {
		q.logger.Errorf("Failed to create event %v: %v.", e.Key, err)
		return e.Key, err
	}
	q.logger.Infof("Event enqueued with key %v, ID %v.", e.Key, e.ID)

	q.processEvent(e)

	return e.Key, nil
}

func (q *PersistentQueue) processEvent(e *Event) {
	q.wg.Add(1)
	respChan := make(chan eventqueue.Response)

	// Ignoring error -- currently only occurs if event fails validation, which
	// we check in Enqueue.
	q.logger.Infof("Enqueuing %v with EventQueue.", e.Key)
	_ = q.EventQueue.Enqueue(e.Event, respChan)

	go func() {
		q.logger.Debugf("Waiting for response for %v.", e.Key)
		resp := <-respChan
		q.logger.Debugf("Received response for %v.", e.Key)

		if resp.Error != nil {
			e.Status = StatusError
			q.logger.Infof("EventQueue returned error for %v: %v, %+v", e.Key, resp.Error, resp.Response)
		} else {
			e.Status = StatusSuccess
			q.logger.Infof("EventQueue returned success for %v. ", e.Key)
		}

		err := e.Update(q.Events)
		if err != nil {
			q.logger.Error(err)
		}
		q.logger.Infof("Set status of %v to %v.", e.Key, e.Status)
		q.wg.Done()
	}()
}
