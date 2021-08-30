package common

import "time"

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func NewClock() Clock {
	c := &realClock{}
	return c
}

func (realClock) Now() time.Time { return time.Now() }
