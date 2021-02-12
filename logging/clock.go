package logging

// We need to mock out the clock for tests; we'll use this to do it.

import "code.cloudfoundry.org/clock"

var currentClock clock.Clock

func initClock() {
	currentClock = clock.NewClock()
}
