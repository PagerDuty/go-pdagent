package eventqueue

import (
	"context"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/eventsapi"
)

var DefaultProcessor = EventProcessor

const MaxRetryTimeout = 30 * time.Second

type Processor func(Job, chan bool)

// EventProcessor is a Job processor for use by an EventQueue specifically
// designed to send and receive from the PagerDuty Events V1 or V2 API.
//
// It accepts a Job containing an EventContainer
func EventProcessor(job Job, stop chan bool) {
	ctx := context.Background()
	resp, err := eventsapi.Enqueue(ctx, job.EventContainer)

	job.ResponseChan <- Response{resp, err}
}
