package eventsapi

type EventVersion int

const (
	EventVersion1 EventVersion = iota
	EventVersion2
)

func (ev EventVersion) String() string {
	return []string{"v1", "v2"}[ev]
}

var StringToEventVersion = map[string]EventVersion{
	"v1": EventVersion1,
	"v2": EventVersion2,
}
