package eventsapi

import (
	"context"
	"net/http"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func mockEndpointV2(statusCode int, response interface{}) *gock.Response {
	mock := gock.New("https://events.pagerduty.com").
		Post("/v2/enqueue").
		Reply(statusCode)

	if response != nil {
		mock = mock.JSON(response)
	}

	return mock
}

func TestEnqueueV2Success(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV2{
		Status:   "success",
		Message:  "Event processed",
		DedupKey: "12345",
	}

	mockEndpointV2(200, mockResponse)

	event := EventV2{
		RoutingKey:  "11863b592c824bfc8989d9cba76abcde",
		EventAction: "trigger",
		Payload: PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
	}

	resp, err := EnqueueV2(context.Background(), http.DefaultClient, event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

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

func TestEnqueueV2InvalidEvent(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV2{
		Status:  "invalid event",
		Message: "Event object is invalid",
		Errors:  []string{"Length of 'routing_key' is incorrect (should be 32 characters)"},
	}

	mockEndpointV2(400, mockResponse)

	event := EventV2{
		RoutingKey:  "11863b592c824bfc",
		EventAction: "trigger",
		Payload: PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
	}

	resp, err := EnqueueV2(context.Background(), http.DefaultClient, event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

	if resp.Status != "invalid event" {
		t.Errorf("Expected status to be \"success\", was \"%v\"", resp.Status)
	}

	if resp.Message != "Event object is invalid" {
		t.Errorf("Expected message to be \"Event processed\", was \"%v\"", resp.Message)
	}

	if resp.Errors[0] != "Length of 'routing_key' is incorrect (should be 32 characters)" {
		t.Errorf("Expected message to be \"Length of 'routing_key' is incorrect (should be 32 characters)\", was \"%v\"", resp.Errors[0])
	}
}

func TestEnqueueV2TooManyRequests(t *testing.T) {
	defer gock.Off()

	mockEndpointV2(429, nil)

	event := EventV2{
		RoutingKey:  "11863b592c824bfc8989d9cba76abcde",
		EventAction: "trigger",
		Payload: PayloadV2{
			Summary:  "PagerDuty Agent `CreateV1` Test",
			Source:   "pdagent",
			Severity: "error",
		},
	}

	resp, err := EnqueueV2(context.Background(), http.DefaultClient, event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

	if resp.StatusCode != 429 {
		t.Errorf("Expected status code to be 429, was %v", resp.Status)
	}
}
