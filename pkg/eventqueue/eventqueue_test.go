package eventqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"testing"
	"time"
)

func TestEventQueueSimple(t *testing.T) {
	eq := NewEventQueue()
	defer eq.Shutdown()
	respChan := make(chan Response)
	event := mockEventV2(common.GenerateKey())

	processor := func(job Job, _ chan bool) {
		if job.Event != &event {
			t.Error("Expected enqueued event to match job event.")
		}

		apiResponse := eventsapi.ResponseV2{
			Status:   "success",
			Message:  "Event processed",
			DedupKey: "12345",
		}

		job.ResponseChan <- Response{Response: &apiResponse}
	}
	eq.Processor = processor

	err := eq.Enqueue(&event, respChan)
	if err != nil {
		t.Error(err)
	}

	resp := <-respChan
	respV2 := resp.Response.(*eventsapi.ResponseV2)
	if respV2.Status != "success" {
		t.Error("Expected successful response.")
	}
}

// For this test we enqueue two events in the same queue (by routing key)
// then add a delay to the first one on the processing side.
//
// The expectation is events will still be processed in order due to being
// processed by a single worker queue.
func TestEventQueueSingleOrdering(t *testing.T) {
	eq := NewEventQueue()
	defer eq.Shutdown()

	key := common.GenerateKey()
	event1 := mockEventV2(key)
	event2 := mockEventV2(key)
	respChan1 := make(chan Response)
	respChan2 := make(chan Response)
	var receivedEvents []eventsapi.Event

	processor := func(job Job, _ chan bool) {
		if job.Event == &event1 {
			time.Sleep(time.Second)
		}

		receivedEvents = append(receivedEvents, job.Event)
		job.ResponseChan <- Response{}
	}
	eq.Processor = processor

	_ = eq.Enqueue(&event1, respChan1)
	_ = eq.Enqueue(&event2, respChan2)
	<-respChan1
	<-respChan2

	if receivedEvents[0] != &event1 {
		t.Error("Expected first event, but instead out of order..")
	}
	if receivedEvents[1] != &event2 {
		t.Error("Expected second event, but instead out of order..")
	}
}

// For this test we enqueue two events in different queues (by routing key)
// then add a delay to the first one on the processing side.
//
// The expectation is that the second (non-delayed) event will be processed
// first as it's now in a separate queue.
func TestEventQueueMultiOrdering(t *testing.T) {
	eq := NewEventQueue()
	defer eq.Shutdown()

	event1 := mockEventV2(common.GenerateKey())
	event2 := mockEventV2(common.GenerateKey())
	event2.RoutingKey = "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY"
	respChan1 := make(chan Response)
	respChan2 := make(chan Response)
	var receivedEvents []eventsapi.Event

	processor := func(job Job, _ chan bool) {
		if job.Event == &event1 {
			time.Sleep(time.Second)
		}

		receivedEvents = append(receivedEvents, job.Event)
		job.ResponseChan <- Response{}
	}
	eq.Processor = processor

	_ = eq.Enqueue(&event1, respChan1)
	_ = eq.Enqueue(&event2, respChan2)
	<-respChan1
	<-respChan2

	if receivedEvents[0] != &event2 {
		t.Error("Expected second event, but instead out of order..")
	}
	if receivedEvents[1] != &event1 {
		t.Error("Expected first event, but instead out of order..")
	}
}
