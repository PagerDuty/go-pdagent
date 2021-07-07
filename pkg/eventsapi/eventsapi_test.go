package eventsapi

import (
	"context"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestCommonEnqueueV2(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV2{
		Status:   "success",
		Message:  "Event processed",
		DedupKey: "12345",
	}

	mockEndpointV2(200, mockResponse)
	gock.InterceptClient(DefaultHTTPClient)

	event := EventContainer{
		EventVersion: EventVersion2,
		EventData: []byte(`
			{
				"routing_key":  "11863b592c824bfc8989d9cba76abcde",
				"event_action": "trigger",
				"payload": {
					"summary":  "PagerDuty Agent CreateV1 Test",
					"source":   "pdagent",
					"severity": "error",
				}
			}
		`),
	}

	vagueResp, err := Enqueue(context.Background(), &event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

	resp := vagueResp.(*ResponseV2)

	if resp.Status != "success" {
		t.Errorf("Expected status to be \"success\", was \"%v\"", resp.Status)
	}

	if resp.Message != "Event processed" {
		t.Errorf("Expected message to be \"Event processed\", was \"%v\"", resp.Message)
	}

	if resp.DedupKey != "12345" {
		t.Errorf("Expected message to be \"12345\", was \"%v\"", resp.DedupKey)
	}
}

func TestCommonEnqueueV1(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV1{
		Status:      "success",
		Message:     "Event processed",
		IncidentKey: "12345",
	}

	mockEndpointV1(200, mockResponse)
	gock.InterceptClient(DefaultHTTPClient)

	event := EventContainer{
		EventVersion: EventVersion1,
		EventData: []byte(`
			{
				"service_key": "11863b592c824bfc8989d9cba76abcde",
				"event_type":  "trigger",
				"description": "PagerDuty Agent CreateV1 Test"
			}
		`),
	}

	vagueResp, err := Enqueue(context.Background(), &event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

	resp := vagueResp.(*ResponseV1)

	if resp.Status != "success" {
		t.Errorf("Expected status to be \"success\", was \"%v\"", resp.Status)
	}

	if resp.Message != "Event processed" {
		t.Errorf("Expected message to be \"Event processed\", was \"%v\"", resp.Message)
	}

	if resp.IncidentKey != "12345" {
		t.Errorf("Expected message to be \"12345\", was \"%v\"", resp.IncidentKey)
	}
}
