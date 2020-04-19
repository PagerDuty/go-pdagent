package eventsapi

import (
	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	"go.uber.org/zap"
	"math"
	"net/http"
	"time"
)

const defaultMaxInterval = 30 * time.Second
const defaultMaxRetries = 10

// RetryTransport provides automatic retry support as an `http.RoundTripper`.
//
// Default cases are when a 429 or 500-series error is encountered, with
// an exponential backoff determined by `Backoff` and a maximum retry count
// of `MaxRetries`.
//
// Example basic usage:
//
//     client :=  &http.Client{
//		   Transport: NewRetryTransport(),
//	   }
//
//     client.Get("https://www.pagerduty.com")
//
type RetryTransport struct {
	MaxRetries  int
	Transport   http.RoundTripper
	Backoff     func(int) time.Duration
	IsRetryable func(*http.Response) bool
	IsSuccess   func(*http.Response) bool

	log *zap.SugaredLogger
}

func NewRetryTransport() RetryTransport {
	return RetryTransport{
		MaxRetries: defaultMaxRetries,
		Transport:  http.DefaultTransport,

		Backoff:     calculateBackoff,
		IsRetryable: isRetryable,
		IsSuccess:   isSuccess,

		log: common.Logger.Named("RetryTransport"),
	}
}

// Implementing the `http.RoundTripper` interface.
func (r RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	ctx := req.Context()

	for tries := 0; tries < r.MaxRetries; tries++ {
		resp, err = r.Transport.RoundTrip(req)

		if r.IsSuccess(resp) || (err != nil && !r.IsRetryable(resp)) {
			r.log.Debugf("Successful or non-retryable response.")
			return resp, err
		} else if !r.IsRetryable(resp) {
			// This branch might seem redundant, but a typical case would be
			// when there's an HTTP error that doesn't result in a client
			// error (e.g. 404).
			r.log.Errorf("Error encountered: %v", resp.Status)
			return resp, ErrAPIError
		}

		backoff := r.Backoff(tries)
		sleep := time.After(backoff)
		r.log.Infof("Retrying job, attempt %v, delay %v", tries+1, backoff)

		select {
		case <-sleep:
			continue
		case <-ctx.Done():
			// The underlying `Transport` should also handle this, but our
			// handling breaks us out of sleep.
			err = ctx.Err()
			break
		}
	}

	if resp != nil {
		r.log.Errorf("Exhausted retries, status was: %v", resp.StatusCode)
	} else {
		r.log.Errorf("Exhausted retries, error was: %v", err)
	}

	// If we exhaust our retries, return the last response and error received.
	return resp, err
}

// calculateBackoff returns an exponential duration based on the try count.
//
// Currently back-off looks like: 1s, 2s, 4s, 8s, 16s, then capping at
// MaxRetryTimeout.
func calculateBackoff(try int) time.Duration {
	duration := time.Duration(math.Pow(2, float64(try))) * time.Second
	if duration > defaultMaxInterval {
		duration = defaultMaxInterval
	}
	return duration
}

// isRetryable returns true if the corresponding request failed but can be
// retried.
//
// Per documentation this is when the there's a network failure or the response
// status code is 429 or a 5XX.
func isRetryable(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	return resp.StatusCode == 429 || resp.StatusCode/100 == 5
}

// isSuccess returns true if the corresponding request was successful.
//
// Per documentation this is when the server responds with a 202, but we treat
// any 2XX as a success.
func isSuccess(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	return resp.StatusCode/100 == 2
}
