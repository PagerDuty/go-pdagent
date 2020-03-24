package persistentqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"testing"
	"time"
)

func TestPersistentQueueSimple(t *testing.T) {
	setup(t)
	defer teardown(t)

	eq := NewMockEventQueue()

	q, err := NewPersistentQueue(tmpDbFile, WithEventQueue(eq))
	if err != nil {
		t.Fatal(err)
	}

	event := eventsapi.EventV2{
		RoutingKey:  "11863b592c824bfc8989d9cba76abcde",
		EventAction: "trigger",
		Payload: eventsapi.PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
	}

	key, err := q.Enqueue(&event)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	persistedEvent, err := FindEventByKey(q.Events, key)
	if err != nil {
		t.Fatal("Could not find persisted event.")
	}

	if persistedEvent.Status != StatusSuccess {
		t.Fatalf("Expected event status to be success, was %v.", persistedEvent.Status)
	}

	q.Shutdown()
}
