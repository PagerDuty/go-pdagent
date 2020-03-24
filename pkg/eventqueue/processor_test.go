package eventqueue

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"gopkg.in/h2non/gock.v1"
	"testing"
	"time"
)

func TestEventsV2ProcessorSimple(t *testing.T) {
	defer gock.Off()

	gock.New("https://events.pagerduty.com").
		Post("/v2/enqueue").
		Reply(200).
		JSON(eventsapi.ResponseV2{
			Status:   "success",
			Message:  "Event processed",
			DedupKey: "12345",
		})
	gock.InterceptClient(eventsapi.DefaultHttpClient)

	respChan := make(chan Response)
	stopChan := make(chan bool)
	event := mockEventV2(common.GenerateKey())

	job := Job{
		Event:        &event,
		ResponseChan: respChan,
		Logger:       common.Logger,
	}

	//
	go func() {
		time.Sleep(30 * time.Second)
		close(stopChan)
	}()

	go EventProcessor(job, stopChan)

	timer := time.After(time.Second)
	select {
	case resp := <-respChan:
		respV2 := resp.Response.(*eventsapi.ResponseV2)

		if resp.Error != nil {
			t.Error("Unexpected failure during job processing.")
		} else if respV2.DedupKey != "12345" {
			t.Error("DedupKey mismatch.")
		}
	case <-timer:
		t.Error("Expected response from processor, none received.")
	}
}

func TestEventsV2ProcessorRetries(t *testing.T) {
	defer gock.Off()

	// Respond twice with 429s, which should be retryable.
	gock.New("https://events.pagerduty.com").
		Times(2).
		Post("/v2/enqueue").
		Reply(429)

	// Then response with a 200, which should trigger success.
	gock.New("https://events.pagerduty.com").
		Post("/v2/enqueue").
		Reply(200).
		JSON(eventsapi.ResponseV2{
			Status:   "success",
			Message:  "Event processed",
			DedupKey: "12345",
		})

	gock.InterceptClient(eventsapi.DefaultHttpClient)

	respChan := make(chan Response)
	stopChan := make(chan bool)
	event := mockEventV2(common.GenerateKey())

	job := Job{
		Event:        &event,
		ResponseChan: respChan,
		Logger:       common.Logger,
	}

	go func() {
		time.Sleep(30 * time.Second)
		close(stopChan)
	}()

	go EventProcessor(job, stopChan)

	timer := time.After(5 * time.Second)
	select {
	case resp := <-respChan:
		respV2 := resp.Response.(*eventsapi.ResponseV2)

		if resp.Error != nil {
			t.Error("Unexpected failure during job processing.")
		} else if respV2.DedupKey != "12345" {
			t.Error("DedupKey mismatch.")
		}
	case <-timer:
		t.Error("Expected response from processor, none received.")
	}
}
