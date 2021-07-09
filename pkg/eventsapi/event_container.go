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
		var v1Event EventV1
		err := json.Unmarshal(ec.EventData, &v1Event)
		return &v1Event, err
	case EventVersion2:
		var v2Event EventV2
		err := json.Unmarshal(ec.EventData, &v2Event)
		return &v2Event, err
	default:
		return nil, ErrUnrecognizedEventType
	}
}
