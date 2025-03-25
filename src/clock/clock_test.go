package clock

import (
	"testing"
)

func TestUpdateClock(t *testing.T) {
	clock = 0
	UpdateClock()
	if clock != 1 {
		t.Errorf("Expected clock to be 1, but got %d", clock)
	}
}
