package eventqueue

import (
	"context"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"math"
	"time"
)

var DefaultProcessor = EventProcessor

const MaxRetryTimeout = 30 * time.Second

type Processor func(Job, chan bool)

// EventProcessor is a Job processor for use by an EventQueue specifically
// designed to send and receive from the PagerDuty Events V2 API.
//
// It accepts a Job containing an Event
func EventProcessor(job Job, stop chan bool) {
	ctx := context.Background()
	resp, err := eventsapi.Enqueue(ctx, job.Event)

	job.ResponseChan <- Response{resp, err}
}

// calculateBackoff returns an exponential duration based on the try count.
//
// Currently back-off looks like: 1s, 2s, 4s, 8s, 16s, then capping at
// MaxRetryTimeout.
func calculateBackoff(try int) time.Duration {
	duration := time.Duration(math.Pow(2, float64(try))) * time.Second
	if duration > MaxRetryTimeout {
		duration = MaxRetryTimeout
	}
	return duration
}
