package clock

import (
	"testing"
)

func TestUpdateClock(t *testing.T) {
	safeClock.clock = 0
	UpdateClock()
	if safeClock.clock != 1 {
		t.Errorf("Expected clock to be 1, but got %d", safeClock.clock)
	}
}
