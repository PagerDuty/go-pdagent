package eventsapi

import (
	"context"
	"net/http"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func mockEndpointV1(statusCode int, response interface{}) *gock.Response {
	mock := gock.New("https://events.pagerduty.com").
		Post("/generic/2010-04-15/create_event.json").
		Reply(statusCode)

	if response != nil {
		mock = mock.JSON(response)
	}

	return mock
}

func TestCreateV1Success(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV1{
		Status:      "success",
		Message:     "Event processed",
		IncidentKey: "12345",
	}

	mockEndpointV1(200, mockResponse)

	event := EventV1{
		ServiceKey:  "11863b592c824bfc8989d9cba76abcde",
		EventType:   "trigger",
		Description: "PagerDuty Agent `CreateV1` Test",
	}

	resp, err := CreateV1(context.Background(), http.DefaultClient, event)
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

	if resp.IncidentKey != "12345" {
		t.Errorf("Expected message to be \"12345\", was \"%v\"", resp.IncidentKey)
	}
}

func TestCreateV1InvalidEvent(t *testing.T) {
	defer gock.Off()

	mockResponse := ResponseV1{
		Status:  "invalid event",
		Message: "Event object is invalid",
		Errors:  []string{"Length of 'routing_key' is incorrect (should be 32 characters)"},
	}

	mockEndpointV1(400, mockResponse)

	event := EventV1{
		ServiceKey:  "11863b592c824bfc",
		EventType:   "trigger",
		Description: "PagerDuty Agent `CreateV1` Test",
	}

	resp, err := CreateV1(context.Background(), http.DefaultClient, event)
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

func TestCreateV1TooManyRequests(t *testing.T) {
	defer gock.Off()

	mockEndpointV1(429, nil)

	event := EventV1{
		ServiceKey:  "11863b592c824bfc8989d9cba76abcde",
		EventType:   "trigger",
		Description: "PagerDuty Agent `CreateV1` Test",
	}

	resp, err := CreateV1(context.Background(), http.DefaultClient, event)
	if err != nil {
		t.Error("Unexpected error during event creation", err)
		return
	}

	if resp.HTTPResponse.StatusCode != 429 {
		t.Errorf("Expected status code to be 429, was %v", resp.Status)
	}
}
