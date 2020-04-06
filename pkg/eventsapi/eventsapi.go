package eventsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/PagerDuty/pagerduty-agent/pkg/common"
)

var ErrInvalidRoutingKey = errors.New("invalid routing key")

// ErrUnrecognizedEventType occurs when an event isn't supported by the events
// API.
var ErrUnrecognizedEventType = errors.New("unrecognized event type")

type Event interface {
	GetRoutingKey() string
	Validate() error
}

// Response defines a minimal interface for the events APIs' HTTP responses.
type Response interface {
	GetHTTPResponse() *http.Response
	IsRetryable() bool
	IsSuccess() bool
	SetHTTPResponse(*http.Response)
	SetRetryable(bool)
}

// BaseResponse is a minimal implementation of the `Response` interface.
type BaseResponse struct {
	HTTPResponse *http.Response
	retryable    bool
}

func (br *BaseResponse) GetHTTPResponse() *http.Response {
	return br.HTTPResponse
}

// IsRetryable returns true if the corresponding request failed but can be
// retried.
//
// Per documentation this is when the there's a network failure or the response
// status code is 429 or a 5XX.
func (br *BaseResponse) IsRetryable() bool {
	if br.HTTPResponse == nil {
		return false
	}
	statusCode := br.HTTPResponse.StatusCode
	return br.retryable || statusCode == 429 || statusCode/100 == 5
}

// IsSuccess returns true if the corresponding request was successful.
//
// Per documentation this is when the server responds with a 202, but we treat
// any 2XX as a success.
func (br *BaseResponse) IsSuccess() bool {
	if br.HTTPResponse == nil {
		return false
	}
	return br.HTTPResponse.StatusCode/100 == 2
}

// SetRetryable indicates explicitly whether a corresponding request was
// retryable.
//
// One particular use case is when a network error occurs -- we don't have a
// HTTP response to analyze but want a way to indicate back to clients that it's
// safe to retry.
func (br *BaseResponse) SetRetryable(retryable bool) {
	br.retryable = retryable
}

// SetHTTPResponse sets `HTTPResponse` on a response.
func (br *BaseResponse) SetHTTPResponse(resp *http.Response) {
	br.HTTPResponse = resp
}

type enqueueConfig struct {
	HTTPClient *http.Client
}

var DefaultHttpClient *http.Client

var defaultEnqueueConfig enqueueConfig

var defaultUserAgent string

func init() {
	DefaultHttpClient = &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   30 * time.Second,
	}

	defaultEnqueueConfig = enqueueConfig{
		HTTPClient: DefaultHttpClient,
	}

	defaultUserAgent = userAgent()
}

type EnqueueOption func(*enqueueConfig)

// WithHTTPClient is an option for use in conjunction with Enqueue allowing
// a user to override our default HTTP client with their own.
func WithHTTPClient(client *http.Client) EnqueueOption {
	return func(ec *enqueueConfig) {
		ec.HTTPClient = client
	}
}

// Enqueue an event to either the V1 or V2 events API depending on event type.
func Enqueue(context context.Context, event Event, options ...EnqueueOption) (Response, error) {
	config := defaultEnqueueConfig
	for _, option := range options {
		option(&config)
	}

	switch e := event.(type) {
	case *EventV1:
		return CreateV1(context, config.HTTPClient, e)
	case *EventV2:
		return EnqueueV2(context, config.HTTPClient, e)
	default:
		return nil, ErrUnrecognizedEventType
	}
}

// enqueueEvent handles common operations around encoding, sending, then
// receiving and decoding from both the V1 and V2 events APIs.
func enqueueEvent(context context.Context, client *http.Client, url string, event Event, response Response) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", defaultUserAgent)
	req.WithContext(context)

	httpResp, err := client.Do(req)
	if err != nil {
		response.SetRetryable(true)
		return err
	}
	response.SetHTTPResponse(httpResp)

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	_ = json.Unmarshal(respBody, &response)

	return nil
}

func userAgent() string {
	version := common.Version
	system := runtime.GOOS
	commit := common.Commit
	date := common.Date

	return fmt.Sprintf("pagerduty-agent/%v (%v, commit: %v, date: %v)", version, system, commit, date)
}

func validateRoutingKey(routingKey string) error {
	if len(routingKey) < 32 {
		return ErrInvalidRoutingKey
	}
	return nil
}
