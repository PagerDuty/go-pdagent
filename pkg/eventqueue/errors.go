package eventqueue

import (
	"errors"
	"fmt"
)

var ErrAPIError = errors.New("an API error was encountered while processing events")

var ErrJobStopped = errors.New("job stopped while retrying")

type ErrBufferOverflow struct {
	key  string
	size int
}

func (e *ErrBufferOverflow) Error() string {
	return fmt.Sprintf("buffer for %v hit limit of %v, normally indicating an excess events", e.key, e.size)
}
