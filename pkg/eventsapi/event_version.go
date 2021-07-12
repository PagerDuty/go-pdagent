package eventsapi

type EventVersion string

var EventVersion1 EventVersion = "v1"
var EventVersion2 EventVersion = "v2"

func (ev EventVersion) String() string {
	return string(ev)
}

var StringToEventVersion = map[string]EventVersion{
	"v1": EventVersion1,
	"v2": EventVersion2,
}
