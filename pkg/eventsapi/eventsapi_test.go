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

	event := EventV2{
		RoutingKey:  "11863b592c824bfc8989d9cba76abcde",
		EventAction: "trigger",
		Payload: PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
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

	event := EventV1{
		ServiceKey:  "11863b592c824bfc8989d9cba76abcde",
		EventType:   "trigger",
		Description: "PagerDuty Agent `CreateV1` Test",
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
