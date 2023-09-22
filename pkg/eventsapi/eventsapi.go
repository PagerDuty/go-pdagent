package eventsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/common"
)

var ErrInvalidRoutingKey = errors.New("invalid routing key")

// ErrUnrecognizedEventType occurs when an event isn't supported by the events
// API.
var ErrUnrecognizedEventType = errors.New("unrecognized event type")

type Event interface {
	GetRoutingKey() string
	Validate() error
	Version() EventVersion
}

// Response defines a minimal interface for the events APIs' HTTP responses.
type Response interface {
	GetHTTPResponse() *http.Response
	SetHTTPResponse(*http.Response)
}

// BaseResponse is a minimal implementation of the `Response` interface.
type BaseResponse struct {
	HTTPResponse *http.Response
	retryable    bool
}

func (br *BaseResponse) GetHTTPResponse() *http.Response {
	return br.HTTPResponse
}

// SetHTTPResponse sets `HTTPResponse` on a response.
func (br *BaseResponse) SetHTTPResponse(resp *http.Response) {
	br.HTTPResponse = resp
}

type enqueueConfig struct {
	HTTPClient *http.Client
}

var DefaultHTTPClient *http.Client

var defaultEnqueueConfig enqueueConfig

var defaultUserAgent string

func init() {
	DefaultHTTPClient = &http.Client{
		Transport: common.NewRetryTransport(),
		Timeout:   5 * time.Minute,
	}

	defaultEnqueueConfig = enqueueConfig{
		HTTPClient: DefaultHTTPClient,
	}

	defaultUserAgent = common.UserAgent()
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
func Enqueue(context context.Context, eventContainer *EventContainer, options ...EnqueueOption) (Response, error) {
	config := defaultEnqueueConfig
	for _, option := range options {
		option(&config)
	}

	event, err := eventContainer.UnmarshalEvent()
	if err != nil {
		return nil, err
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
	req = req.WithContext(context)

	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	response.SetHTTPResponse(httpResp)

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	_ = json.Unmarshal(respBody, &response)
	if common.IsSuccessResponse(httpResp, err) {
		return nil
	}

	return ErrAPIError
}

func validateRoutingKey(routingKey string) error {
	if len(routingKey) < 32 {
		return ErrInvalidRoutingKey
	}
	return nil
}
