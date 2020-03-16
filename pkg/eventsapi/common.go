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

// enqueueEvent handles common operations around encoding, sending, then
// receiving and decoding from both the V1 and V2 events APIs.
func enqueueEvent(context context.Context, client *http.Client, url string, event interface{}, response interface{}) (*http.Response, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.WithContext(context)

	httpResp, err := client.Do(req)
	if err != nil {
		return httpResp, err
	}

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return httpResp, err
	}

	_ = json.Unmarshal(respBody, &response)

	return httpResp, nil
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
