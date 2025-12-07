package storage

import "time"

type TimeEntry struct {
	ID int64
	ProjectName string
	StartTime time.Time
	EndTime* time.Time
	Description string
}

func (t* TimeEntry) Duration() time.Duration {
	if( t.EndTime == nil) {
		return time.Since(t.StartTime)
	}

	return t.EndTime.Sub(t.StartTime)
}

func (t* TimeEntry) IsRunning() bool {
	return t.EndTime == nil
}