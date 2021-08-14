package undirect

import (
	"time"
)

// clock is the interface for implementing clock, the implementation can achieve in different ways in terms of
//
// 1. lamport clock,
// 2. vector clock,
// 3. ntp
// ...
//
// e.g.
//
// type ntpClock struct {
// 	addr string
// }
//
// func NewNTP(addr string) Clock {
// 	return &ntpClock{addr}
// }
//
// func (c *ntpClock) Now() time.Time {
// 	time, err := ntp.Now(ntpClock.addr)
//  if err != nil {
// 		return time.Now()
//  }
// 	return time
// }
type Clock interface {
	// function to return now that define by the implementation
	Now() time.Time
}

type clock struct{}

func (c *clock) Now() time.Time { return time.Now() }

// test clock use for mock the time for unit test and return a unix time that start with same timestamp
type testCkock struct {
	now *time.Time
}

func (tc *testCkock) Now() time.Time {
	now := time.Unix(0, 0)
	if tc.now == nil {
		tc.now = &now
	}
	return *tc.now
}

// add duration for the time of the test clock to set the timeline for unit test
func (tc *testCkock) AddDuration(d time.Duration) {
	now := tc.Now().Add(d)
	tc.now = &now
	return
}

// Sync the clocks, as physical clock in real world will never in sync,
// there should be any of:
//
// 1. lamport clock,
// 2. vector clock,
// 3. version vector,
// 4. true time,
// 5. ntp,
// 6. hlc,
// 7. single point of true solution,
//
// for solving the time difference for sequential data in different server,
// right now just sync up the clocks without using these ways.
func (tc *testCkock) SyncWith(with *testCkock) {
	if tc.now.Before(*with.now) {
		now := *with.now
		tc.now = &now
		return
	}
	now := *tc.now
	with.now = &now
}
