package todo

import "time"

// Clock provides time functionality for testability
type Clock interface {
	Now() time.Time
}

// RealClock uses actual system time
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// TestClock is a mock clock for deterministic testing
type TestClock struct {
	current time.Time
}

func NewTestClock(start time.Time) *TestClock {
	return &TestClock{current: start}
}

func (tc *TestClock) Now() time.Time {
	return tc.current
}

func (tc *TestClock) Advance(d time.Duration) {
	tc.current = tc.current.Add(d)
}
