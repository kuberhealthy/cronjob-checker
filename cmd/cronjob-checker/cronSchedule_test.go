package main

import (
	"testing"
	"time"
)

// TestScheduleWindow verifies the schedule window bounds.
func TestScheduleWindow(t *testing.T) {
	// Use a fixed reference time.
	reference := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)

	// Build a window around the reference.
	lower, upper := scheduleWindow(reference, time.Minute*10)

	// Validate the bounds are symmetric around the reference.
	if lower != reference.Add(-5*time.Minute) {
		t.Fatalf("expected lower bound to be -5m, got %s", lower)
	}
	if upper != reference.Add(5*time.Minute) {
		t.Fatalf("expected upper bound to be +5m, got %s", upper)
	}
}
