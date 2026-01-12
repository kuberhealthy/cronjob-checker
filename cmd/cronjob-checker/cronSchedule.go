package main

import (
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	// scheduleWindowSize defines the acceptable schedule window.
	scheduleWindowSize = time.Minute * 10
)

// findLastCronRunTime computes the last expected run time for a cron schedule.
func findLastCronRunTime(schedule string) time.Time {
	// Parse the cron schedule expression.
	cronSchedule := cronexpr.MustParse(schedule)

	// Walk forward from one year ago to find the last scheduled time.
	oneYear := time.Hour * 24 * 366
	oneYearAgo := time.Now().Add(-oneYear)
	now := time.Now()
	timeMarker := oneYearAgo

	for {
		nextRunTime := cronSchedule.Next(timeMarker)
		if nextRunTime.After(now) {
			return timeMarker
		}
		timeMarker = nextRunTime
	}
}

// scheduleWindow builds the acceptable schedule window around a target time.
func scheduleWindow(target time.Time, window time.Duration) (time.Time, time.Time) {
	// Build bounds for the window.
	lowerBound := target.Add(-1 * (window / 2))
	upperBound := target.Add(window / 2)

	return lowerBound, upperBound
}
