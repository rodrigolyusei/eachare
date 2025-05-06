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

func TestUpdateMaxClock(t *testing.T) {
	safeClock.clock = 0
	UpdateMaxClock(10)
	if safeClock.clock != 11 {
		t.Errorf("Expected clock to be 11, but got %d", safeClock.clock)
	}
}

func TestGetClock(t *testing.T) {
	safeClock.clock = 10
	value := GetClock()
	if value != 10 {
		t.Errorf("Expected value to be 10, but got %d", value)
	}
}
