package eventqueue

import (
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
	"github.com/PagerDuty/go-pdagent/test"
	"gopkg.in/h2non/gock.v1"
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
	gock.InterceptClient(eventsapi.DefaultHTTPClient)

	respChan := make(chan Response)
	stopChan := make(chan bool)
	event := test.MockEventContainerV2(common.GenerateKey())

	job := Job{
		EventContainer: &event,
		ResponseChan:   respChan,
		Logger:         common.Logger,
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

func TestEventsV2ProcessorError(t *testing.T) {
	defer gock.Off()

	gock.New("https://events.pagerduty.com").
		Post("/v2/enqueue").
		Reply(400).
		JSON(eventsapi.ResponseV2{
			Status:  "invalid event",
			Message: "Event object is invalid",
			Errors: []string{
				"'payload.severity' is invalid (must be one of the following: 'critical', 'warning', 'error' or 'info')",
			},
		})
	gock.InterceptClient(eventsapi.DefaultHTTPClient)

	respChan := make(chan Response)
	stopChan := make(chan bool)
	event := test.MockEventContainerV2(common.GenerateKey())

	job := Job{
		EventContainer: &event,
		ResponseChan:   respChan,
		Logger:         common.Logger,
	}

	go func() {
		time.Sleep(30 * time.Second)
		close(stopChan)
	}()

	go EventProcessor(job, stopChan)

	timer := time.After(time.Second)
	select {
	case resp := <-respChan:
		respV2 := resp.Response.(*eventsapi.ResponseV2)

		if resp.Error == nil {
			t.Error("Expected an error in the response.")
		}
		if respV2.Status != "invalid event" {
			t.Error("Invalid response status")
		}
	case <-timer:
		t.Error("Expected response from processor, none received.")
	}
}
