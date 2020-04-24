package eventsapi

import (
	"bytes"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"testing"
	"time"
)

func mockEventV2(key string) EventV2 {
	return EventV2{
		RoutingKey:  key,
		EventAction: "trigger",
		Payload: PayloadV2{
			Summary:  "Test summary",
			Source:   "Test source",
			Severity: "Error",
		},
	}
}

func TestRetryTransportSuccess(t *testing.T) {
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
		JSON(ResponseV2{
			Status:   "success",
			Message:  "Event processed",
			DedupKey: "12345",
		})

	transport := NewRetryTransport()
	transport.Transport = gock.NewTransport()
	transport.Backoff = func(_ int) time.Duration { return time.Millisecond }

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewBuffer([]byte("Hello")))
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected a success response, response was %+v.", resp)
	}
}

func TestRetryTransportLimited(t *testing.T) {
	defer gock.Off()

	// Respond twice with 429s, which should be retryable.
	gock.New("https://events.pagerduty.com").
		Times(defaultMaxRetries).
		Post("/v2/enqueue").
		Reply(429)

	transport := NewRetryTransport()
	transport.Transport = gock.NewTransport()
	transport.Backoff = func(_ int) time.Duration { return time.Millisecond }

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewBuffer([]byte("Hello")))
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if resp.StatusCode != 429 {
		t.Errorf("Expected a too-many-requests response, response was %+v.", resp)
	}
}
