package eventqueue

import "github.com/PagerDuty/go-pdagent/pkg/eventsapi"

func mockEventV2(key string) eventsapi.EventV2 {
	return eventsapi.EventV2{
		RoutingKey:  key,
		EventAction: "trigger",
		Payload: eventsapi.PayloadV2{
			Summary:  "Test summary",
			Source:   "Test source",
			Severity: "Error",
		},
	}
}
