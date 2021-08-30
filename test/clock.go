package test

import "time"

type TestClock struct{}

func (TestClock) Now() time.Time { return time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC) }
