package eventsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// ErrUnrecognizedEventType occurs when an event isn't supported by the events
// API.
var ErrUnrecognizedEventType = errors.New("unrecognized event type")

// Response defines a minimal interface for the events APIs' HTTP responses.
type Response interface {
	SetHTTPResponse(*http.Response)
}

// BaseResponse is a minimal implementation of the `Response` interface.
type BaseResponse struct {
	HTTPResponse *http.Response
}

// SetHTTPResponse sets `HTTPResponse` on a response.
func (br *BaseResponse) SetHTTPResponse(resp *http.Response) {
	br.HTTPResponse = resp
}

// Enqueue an event to either the V1 or V2 events API depending on event type.
func Enqueue(context context.Context, client *http.Client, event interface{}) (interface{}, error) {
	switch e := event.(type) {
	case EventV1:
		return CreateV1(context, client, e)
	case EventV2:
		return EnqueueV2(context, client, e)
	default:
		return nil, ErrUnrecognizedEventType
	}
}

// enqueueEvent handles common operations around encoding, sending, then
// receiving and decoding from both the V1 and V2 events APIs.
func enqueueEvent(context context.Context, client *http.Client, url string, event interface{}, response Response) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.WithContext(context)

	httpResp, err := client.Do(req)
	if err != nil {
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
