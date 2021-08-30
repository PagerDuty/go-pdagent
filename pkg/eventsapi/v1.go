package eventsapi

import (
	"context"
	"net/http"

	"github.com/PagerDuty/go-pdagent/pkg/common"
)

const endpointV1 = "/generic/2010-04-15/create_event.json"

// EventV1 corresponds to a V1 event object.
type EventV1 struct {
	ServiceKey  string       `json:"service_key"`
	EventType   string       `json:"event_type"`
	IncidentKey string       `json:"incident_key,omitempty"`
	Description string       `json:"description"`
	Details     DetailsV1    `json:"details,omitempty"`
	Client      string       `json:"client,omitempty"`
	ClientURL   string       `json:"client_url,omitempty"`
	Contexts    []ContextV1  `json:"contexts,omitempty"`
	Agent       AgentContext `json:"agent,omitempty"`
}

func (e *EventV1) GetRoutingKey() string {
	return e.ServiceKey
}

func (e *EventV1) Validate() error {
	if err := validateRoutingKey(e.ServiceKey); err != nil {
		return err
	}

	return nil
}

func (e *EventV1) Version() EventVersion {
	return EventVersion1
}

// DetailsV1 corresponds to a V1 details object.
type DetailsV1 map[string]interface{}

// AgentContext is used for pdagent integrations
type AgentContext struct {
	QueuedBy string `json:"queued_by"`
	QueuedAt string `json:"queued_at"`
	AgentId  string `json:"agent_id"`
}

// ContextV1 corresponds to a V1 context object.
//
// Technically this can either be a `link` or `image` context type, but
// currently representing as a single type for convenience.
type ContextV1 struct {
	Type   string `json:"type"`
	Href   string `json:"href"`
	Text   string `json:"text,omitempty"`
	Source string `json:"src"`
	Alt    string `json:"alt,omitempty"`
}

// ResponseV1 corresponds to a V1 response.
type ResponseV1 struct {
	BaseResponse

	Status      string   `json:"status,omitempty"`
	Message     string   `json:"message,omitempty"`
	IncidentKey string   `json:"incident_key,omitempty"`
	Errors      []string `json:"errors,omitempty"`
}

// CreateV1 sends an event to explicitly the Events API V1.
//
// Keeping the `create` semantics versus `enqueue` to more closely match the
// service's own.
func CreateV1(context context.Context, client *http.Client, event *EventV1) (*ResponseV1, error) {
	var response ResponseV1
	url := common.PdEventsUrl() + endpointV1
	err := enqueueEvent(context, client, url, event, &response)
	return &response, err
}
