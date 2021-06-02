package common

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"
)

func TestRetryTransportSuccess(t *testing.T) {
	defer gock.Off()

	// Respond twice with 429s, which should be retryable.
	gock.New("https://events.pagerduty.com").
		Times(2).
		Post("/test").
		Reply(429)

	// Then response with a 200, which should trigger success.
	gock.New("https://events.pagerduty.com").
		Post("/test").
		Reply(200).
		JSON("reply")

	transport := NewRetryTransport()
	transport.Transport = gock.NewTransport()
	transport.Backoff = func(_ int) time.Duration { return time.Millisecond }

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Post("https://events.pagerduty.com/test", "application/json", bytes.NewBuffer([]byte("Hello")))
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
		Post("/test").
		Reply(429)

	transport := NewRetryTransport()
	transport.Transport = gock.NewTransport()
	transport.Backoff = func(_ int) time.Duration { return time.Millisecond }

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Post("https://events.pagerduty.com/test", "application/json", bytes.NewBuffer([]byte("Hello")))
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if resp.StatusCode != 429 {
		t.Errorf("Expected a too-many-requests response, response was %+v.", resp)
	}
}
