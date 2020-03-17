package eventsapi

import (
	"context"
	"net/http"
)

const endpointV2 = "https://events.pagerduty.com/v2/enqueue"

// EventV2 corresponds to a V2 event object.
type EventV2 struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	DedupKey    string    `json:"dedup_key,omitempty"`
	Payload     PayloadV2 `json:"payload"`
	Images      []ImageV2 `json:"images,omitempty"`
	Links       []LinkV2  `json:"links,omitempty"`
}

// PayloadV2 corresponds to a V2 payload object.
type PayloadV2 struct {
	Summary       string                 `json:"summary"`
	Source        string                 `json:"source"`
	Severity      string                 `json:"severity"`
	Timestamp     string                 `json:"timestamp,omitempty"`
	Component     string                 `json:"component,omitempty"`
	Group         string                 `json:"group,omitempty"`
	Class         string                 `json:"class,omitempty"`
	CustomDetails map[string]interface{} `json:"custom_details,omitempty"`
}

// ImageV2 corresponds to a V2 image object.
type ImageV2 struct {
	Source string `json:"src"`
	Href   string `json:"href,omitempty"`
	Alt    string `json:"alt,omitempty"`
}

// LinkV2 corresponds to a V2 link object.
type LinkV2 struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

// ResponseV2 corresponds to a V2 response object.
type ResponseV2 struct {
	BaseResponse

	Status   string   `json:"status,omitempty"`
	Message  string   `json:"message,omitempty"`
	DedupKey string   `json:"dedupkey,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

// EnqueueV2 sends an event explicitly to the Events API V2.
func EnqueueV2(context context.Context, client *http.Client, event EventV2) (*ResponseV2, error) {
	response := new(ResponseV2)
	err := enqueueEvent(context, client, endpointV2, event, response)
	return response, err
}
