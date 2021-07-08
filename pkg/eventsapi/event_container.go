package eventsapi

import (
	"encoding/json"
)

type EventContainer struct {
	EventVersion EventVersion
	EventData    json.RawMessage
}

func (ec *EventContainer) UnmarshalEvent() (Event, error) {
	jsonEventData, err := ec.EventData.MarshalJSON()
	if err != nil {
		return nil, err
	}

	switch ec.EventVersion {
	case EventVersion1:
		var v1Event EventV1
		err = json.Unmarshal(jsonEventData, &v1Event)
		return &v1Event, err
	case EventVersion2:
		var v2Event EventV2
		_ = json.Unmarshal(jsonEventData, &v2Event)
		return &v2Event, err
	default:
		return nil, ErrUnrecognizedEventType
	}
}
