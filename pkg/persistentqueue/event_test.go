package persistentqueue

import (
	"testing"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/asdine/storm"
)

func TestEvent(t *testing.T) {
	setup(t)
	defer teardown(t)

	db, err := storm.Open(tmpDbFile)
	if err != nil {
		t.Fatal(err)
	}

	eventsDb := db.From("events")

	genericEvent := eventsapi.EventContainer{
		EventVersion: eventsapi.EventVersion2,
		EventData: map[string]interface{}{
			"routing_key":  "11863b592c824bfc8989d9cba76abcde",
			"event_action": "trigger",
			"payload": map[string]interface{}{
				"summary":  "PagerDuty Agent `CreateV1` Test",
				"source":   "pdagent",
				"severity": "error",
			},
		},
	}

	event, err := NewEvent(&genericEvent)
	if err != nil {
		t.Fatal(err)
	}

	if err = event.Create(eventsDb); err != nil {
		t.Fatal(err)
	}

	retrievedEvent, err := FindEventByKey(eventsDb, event.Key)
	if err != nil {
		t.Fatal(err)
	}

	if event.ID != retrievedEvent.ID {
		t.Fatal("Expected event IDs to match.")
	}

	event.Status = StatusSuccess

	if err = event.Update(eventsDb); err != nil {
		t.Fatal(err)
	}

	retrievedEvent, err = FindEventByKey(eventsDb, event.Key)
	if err != nil {
		t.Fatal(err)
	}

	if retrievedEvent.Status != StatusSuccess {
		t.Fatal("Expected event status to be updated from the DB.")
	}
}
