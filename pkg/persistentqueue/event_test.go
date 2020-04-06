package persistentqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"github.com/asdine/storm"
	"testing"
)

func TestEvent(t *testing.T) {
	setup(t)
	defer teardown(t)

	db, err := storm.Open(tmpDbFile)
	if err != nil {
		t.Fatal(err)
	}

	eventsDb := db.From("events")

	eventAPIEvent := eventsapi.EventV2{
		RoutingKey:  "11863b592c824bfc8989d9cba76abcde",
		EventAction: "trigger",
		Payload: eventsapi.PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
	}

	event, err := NewEvent(&eventAPIEvent)
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
