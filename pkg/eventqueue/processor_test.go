package eventqueue

import (
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
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
