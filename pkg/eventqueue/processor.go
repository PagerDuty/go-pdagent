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
	logger := job.Logger
	c := job.ResponseChan

	// Continues retrying "forever" if an event is retryable, rather than
	// erroring out as we prioritize ensuring eventual ordered delivery.
	tries := 0
	for {
		ctx := context.Background()

		resp, err := eventsapi.Enqueue(ctx, job.Event)

		if err != nil && !resp.IsRetryable() {
			c <- Response{Error: err}
			return
		} else if resp.IsSuccess() {
			c <- Response{resp, nil}
			return
		} else if !resp.IsRetryable() {
			// This branch might seem redundant, but a typical case would be
			// when there's an HTTP error that doesn't result in a client
			// error (e.g. 404).
			c <- Response{resp, ErrAPIError}
			return
		}

		backoff := calculateBackoff(tries)
		sleep := time.After(backoff)
		logger.Infof("Retrying job, attempt %v, delay %v", tries+1, backoff)

		select {
		case <-sleep:
			tries++
		case <-stop:
			logger.Info("Job stopped while retrying.")
			c <- Response{resp, ErrJobStopped}
			return
		}
	}
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
