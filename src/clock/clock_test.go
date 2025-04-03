package clock

import (
	"testing"
)

func TestUpdateClock(t *testing.T) {
	safeClock.Clock = 0
	UpdateClock()
	if safeClock.Clock != 1 {
		t.Errorf("Expected clock to be 1, but got %d", safeClock.Clock)
	}
}
