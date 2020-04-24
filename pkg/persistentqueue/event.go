package persistentqueue

import (
	"errors"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/asdine/storm"
)

const StatusPending = "pending"
const StatusError = "error"
const StatusSuccess = "success"

var ErrV2Only = errors.New("expected a pointer to an Events API V2 event")

// Event represents an queued or processed event.
type Event struct {
	ID           int    `storm:"id,increment"`
	Key          string `storm:"index"`
	RoutingKey   string `storm:"index"`
	Status       string `storm:"index"`
	Event        *eventsapi.EventV2
	ResponseBody []byte
	CreatedAt    time.Time `storm:"index"`
	UpdatedAt    time.Time `storm:"index"`
}

func NewEvent(event eventsapi.Event) (*Event, error) {
	ev2, ok := event.(*eventsapi.EventV2)
	if !ok {
		return nil, ErrV2Only
	}

	return &Event{
		Key:        common.GenerateKey(),
		RoutingKey: event.GetRoutingKey(),
		Status:     StatusPending,
		Event:      ev2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// Create an event within the specified Queue.
//
// Main convenience is ensuring that CreatedAt and UpdatedAt are set.
func (e *Event) Create(db storm.Node) error {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	return db.Save(e)
}

// Update an event within the specified Queue.
//
// Main convenience is ensuring that UpdatedAt is updated.
func (e *Event) Update(db storm.Node) error {
	e.UpdatedAt = time.Now()
	return db.Update(e)
}

func FindEventByKey(db storm.Node, key string) (*Event, error) {
	var event Event
	err := db.One("Key", key, &event)
	return &event, err
}
