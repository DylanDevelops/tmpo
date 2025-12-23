package storage

import "time"

// TimeEntry represents a recorded period of work on a project.
// It includes a unique identifier, the project name, the start time,
// an optional end time (nil indicates the entry is still in progress),
// a free-form description of the work performed, and an optional hourly rate
// (nil indicates no rate was configured when the entry was created).
type TimeEntry struct {
	ID int64
	ProjectName string
	StartTime time.Time
	EndTime *time.Time
	Description string
	HourlyRate *float64
}

// Duration returns the elapsed time for the TimeEntry.
// If EndTime is non-nil, it returns the difference EndTime.Sub(StartTime).
// If EndTime is nil (the entry is ongoing), it returns time.Since(StartTime).
func (t *TimeEntry) Duration() time.Duration {
	if( t.EndTime == nil) {
		return time.Since(t.StartTime)
	}

	return t.EndTime.Sub(t.StartTime)
}

// IsRunning reports whether the TimeEntry is currently running.
// It returns true when EndTime is nil, indicating no end timestamp has been set.
func (t *TimeEntry) IsRunning() bool {
	return t.EndTime == nil
}

// RoundedDuration returns the duration rounded to the nearest minute.
// This is useful for financial calculations where invoicing is typically done
// in minute increments rather than to the second or millisecond.
func (t *TimeEntry) RoundedDuration() time.Duration {
	duration := t.Duration()
	minutes := duration.Round(time.Minute)
	return minutes
}