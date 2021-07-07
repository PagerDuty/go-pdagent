package eventsapi

import (
	"encoding/json"
)

type GenericEvent struct {
	EventVersion EventVersion
	EventData    map[string]interface{}
}

func (ge GenericEvent) Validate() error {
	if ge.EventVersion == EventVersion1 {
		event := ge.getSpecificV1EventStruct()
		return event.Validate()
	} else {
		event := ge.getSpecificV2EventStruct()
		return event.Validate()
	}
}

func (ge GenericEvent) GetRoutingKey() string {
	if ge.EventVersion == EventVersion1 {
		event := ge.getSpecificV1EventStruct()
		return event.GetRoutingKey()
	} else {
		event := ge.getSpecificV2EventStruct()
		return event.GetRoutingKey()
	}
}

func (ge GenericEvent) Version() EventVersion {
	return ge.EventVersion
}

func (ge GenericEvent) getSpecificV1EventStruct() EventV1 {
	jsonV1Event, _ := json.Marshal(ge.EventData)
	var v1Event EventV1
	_ = json.Unmarshal(jsonV1Event, &v1Event)
	return v1Event
}

func (ge GenericEvent) getSpecificV2EventStruct() EventV2 {
	jsonV2Event, _ := json.Marshal(ge.EventData)
	var v2Event EventV2
	_ = json.Unmarshal(jsonV2Event, &v2Event)
	return v2Event
}
