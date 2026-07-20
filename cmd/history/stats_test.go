package history

import (
	"testing"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/stretchr/testify/assert"
)

func entryWithRate(project string, hours float64, rate *float64) *storage.TimeEntry {
	start := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Duration(hours * float64(time.Hour)))

	return &storage.TimeEntry{
		ProjectName: project,
		StartTime:   start,
		EndTime:     &end,
		HourlyRate:  rate,
	}
}

func TestBuildStatsOutput(t *testing.T) {
	t.Run("aggregates totals and sorts projects alphabetically", func(t *testing.T) {
		entries := []*storage.TimeEntry{
			entryWithRate("zeta", 1, nil),
			entryWithRate("alpha", 3, nil),
			entryWithRate("alpha", 0, nil),
		}

		out := buildStatsOutput(entries, "Today", nil)

		assert.Equal(t, "Today", out.Period)
		assert.Equal(t, 4.0, out.TotalHours)
		assert.Equal(t, 3, out.TotalEntries)
		assert.Nil(t, out.ProjectsTracked)
		assert.Nil(t, out.TotalEarnings)
		assert.Empty(t, out.Currency)

		// Sorted: alpha (3h) before zeta (1h)
		assert.Len(t, out.ByProject, 2)
		assert.Equal(t, "alpha", out.ByProject[0].Project)
		assert.Equal(t, 3.0, out.ByProject[0].Hours)
		assert.Equal(t, 75.0, out.ByProject[0].Percentage)
		assert.Equal(t, "zeta", out.ByProject[1].Project)
		assert.Equal(t, 25.0, out.ByProject[1].Percentage)
	})

	t.Run("rounds percentage to one decimal", func(t *testing.T) {
		entries := []*storage.TimeEntry{
			entryWithRate("a", 1, nil),
			entryWithRate("b", 1, nil),
			entryWithRate("c", 1, nil),
		}

		out := buildStatsOutput(entries, "Today", nil)

		// 1/3 => 33.333... rounded to 33.3
		for _, p := range out.ByProject {
			assert.Equal(t, 33.3, p.Percentage)
		}
	})

	t.Run("includes earnings and currency when rates are present", func(t *testing.T) {
		rate := 100.0
		entries := []*storage.TimeEntry{
			entryWithRate("billed", 2, &rate),
			entryWithRate("unbilled", 1, nil),
		}

		out := buildStatsOutput(entries, "Today", nil)

		assert.NotNil(t, out.TotalEarnings)
		assert.Equal(t, 200.0, *out.TotalEarnings)
		assert.NotEmpty(t, out.Currency)

		// Only the billed project carries earnings
		assert.Equal(t, "billed", out.ByProject[0].Project)
		assert.NotNil(t, out.ByProject[0].Earnings)
		assert.Equal(t, 200.0, *out.ByProject[0].Earnings)
		assert.Equal(t, "unbilled", out.ByProject[1].Project)
		assert.Nil(t, out.ByProject[1].Earnings)
	})

	t.Run("passes through projects tracked for all-time view", func(t *testing.T) {
		count := 5
		out := buildStatsOutput(nil, "All Time", &count)

		assert.NotNil(t, out.ProjectsTracked)
		assert.Equal(t, 5, *out.ProjectsTracked)
	})

	t.Run("empty entries produce a zeroed object with a non-nil breakdown", func(t *testing.T) {
		out := buildStatsOutput([]*storage.TimeEntry{}, "Today", nil)

		assert.Equal(t, 0.0, out.TotalHours)
		assert.Equal(t, 0, out.TotalEntries)
		assert.Nil(t, out.TotalEarnings)
		assert.NotNil(t, out.ByProject)
		assert.Len(t, out.ByProject, 0)
	})
}
