package eventsapi

import (
	"encoding/json"
)

type EventContainer struct {
	EventVersion EventVersion
	EventData    json.RawMessage
}

func (ec *EventContainer) UnmarshalEvent() (Event, error) {
	switch ec.EventVersion {
	case EventVersion1:
		var event EventV1
		err := json.Unmarshal(ec.EventData, &event)
		return &event, err
	case EventVersion2:
		var event EventV2
		err := json.Unmarshal(ec.EventData, &event)
		return &event, err
	default:
		return nil, ErrUnrecognizedEventType
	}
}
