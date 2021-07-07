package persistentqueue

import (
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
)

func TestPersistentQueueSimple(t *testing.T) {
	setup(t)
	defer teardown(t)

	eq := NewMockEventQueue()

	q := NewPersistentQueue(WithEventQueue(eq))

	err := q.Start()
	if err != nil {
		t.Fatal("Error starting persistent queue.")
	}

	eventContainer := eventsapi.EventContainer{
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

	key, err := q.Enqueue(&eventContainer)
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

	_ = q.Shutdown()
}
