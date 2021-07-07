package eventsapi

import (
	"encoding/json"
)

type EventContainer struct {
	EventVersion EventVersion
	EventData    map[string]interface{}
}

func (ec *EventContainer) GetEvent() Event {
	jsonEvent, _ := json.Marshal(ec.EventData)
	if ec.EventVersion == EventVersion1 {
		var v1Event EventV1
		_ = json.Unmarshal(jsonEvent, &v1Event)
		return &v1Event
	} else {
		var v2Event EventV2
		_ = json.Unmarshal(jsonEvent, &v2Event)
		return &v2Event
	}
}
