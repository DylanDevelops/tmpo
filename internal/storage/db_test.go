package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *Database {
	db, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			description TEXT,
			hourly_rate REAL
		)
	`)
	assert.NoError(t, err)

	return &Database{db: db}
}

func TestCreateEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		projectName string
		description string
		hourlyRate  *float64
	}{
		{
			name:        "entry without rate",
			projectName: "test-project",
			description: "test description",
			hourlyRate:  nil,
		},
		{
			name:        "entry with rate",
			projectName: "paid-project",
			description: "billable work",
			hourlyRate:  floatPtr(150.0),
		},
		{
			name:        "entry without description",
			projectName: "quick-task",
			description: "",
			hourlyRate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := db.CreateEntry(tt.projectName, tt.description, tt.hourlyRate)

			assert.NoError(t, err)
			assert.NotNil(t, entry)
			assert.Greater(t, entry.ID, int64(0))
			assert.Equal(t, tt.projectName, entry.ProjectName)
			assert.Equal(t, tt.description, entry.Description)
			assert.Nil(t, entry.EndTime)

			if tt.hourlyRate != nil {
				assert.NotNil(t, entry.HourlyRate)
				assert.Equal(t, *tt.hourlyRate, *entry.HourlyRate)
			} else {
				assert.Nil(t, entry.HourlyRate)
			}
		})
	}
}

func TestCreateManualEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now().Add(-1 * time.Hour)
	rate := 100.0

	entry, err := db.CreateManualEntry("manual-project", "manual work", startTime, endTime, &rate)

	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "manual-project", entry.ProjectName)
	assert.Equal(t, "manual work", entry.Description)
	assert.NotNil(t, entry.EndTime)
	assert.WithinDuration(t, startTime, entry.StartTime, time.Second)
	assert.WithinDuration(t, endTime, *entry.EndTime, time.Second)
	assert.Equal(t, rate, *entry.HourlyRate)
}

func TestGetRunningEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// No running entry initially
	running, err := db.GetRunningEntry()
	assert.NoError(t, err)
	assert.Nil(t, running)

	// Create a running entry
	entry, err := db.CreateEntry("test-project", "test", nil)
	assert.NoError(t, err)

	// Should return the running entry
	running, err = db.GetRunningEntry()
	assert.NoError(t, err)
	assert.NotNil(t, running)
	assert.Equal(t, entry.ID, running.ID)
	assert.Nil(t, running.EndTime)

	// Stop the entry
	err = db.StopEntry(entry.ID)
	assert.NoError(t, err)

	// No running entry after stopping
	running, err = db.GetRunningEntry()
	assert.NoError(t, err)
	assert.Nil(t, running)
}

func TestStopEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry, err := db.CreateEntry("test-project", "test", nil)
	assert.NoError(t, err)
	assert.Nil(t, entry.EndTime)

	err = db.StopEntry(entry.ID)
	assert.NoError(t, err)

	// Verify entry was stopped
	stopped, err := db.GetEntry(entry.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stopped.EndTime)
}

func TestGetEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rate := 75.5
	created, err := db.CreateEntry("test-project", "test description", &rate)
	assert.NoError(t, err)

	// Get the entry
	entry, err := db.GetEntry(created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, created.ID, entry.ID)
	assert.Equal(t, "test-project", entry.ProjectName)
	assert.Equal(t, "test description", entry.Description)
	assert.Equal(t, rate, *entry.HourlyRate)

	// Non-existent entry
	_, err = db.GetEntry(9999)
	assert.Error(t, err)
}

func TestGetEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create multiple entries
	for i := 0; i < 5; i++ {
		_, err := db.CreateEntry("test-project", "", nil)
		assert.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get all entries
	entries, err := db.GetEntries(0)
	assert.NoError(t, err)
	assert.Len(t, entries, 5)

	// Entries should be sorted by start_time DESC (newest first)
	for i := 0; i < len(entries)-1; i++ {
		assert.True(t, entries[i].StartTime.After(entries[i+1].StartTime))
	}

	// Get limited entries
	entries, err = db.GetEntries(3)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestGetEntriesByProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create entries for different projects
	_, err := db.CreateEntry("project-a", "task 1", nil)
	assert.NoError(t, err)
	_, err = db.CreateEntry("project-b", "task 2", nil)
	assert.NoError(t, err)
	_, err = db.CreateEntry("project-a", "task 3", nil)
	assert.NoError(t, err)

	// Get entries for project-a
	entries, err := db.GetEntriesByProject("project-a")
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	for _, entry := range entries {
		assert.Equal(t, "project-a", entry.ProjectName)
	}

	// Get entries for project-b
	entries, err = db.GetEntriesByProject("project-b")
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "project-b", entries[0].ProjectName)

	// Get entries for non-existent project
	entries, err = db.GetEntriesByProject("non-existent")
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestGetEntriesByDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// Create entries with different start times
	_, err := db.CreateManualEntry("project", "old", twoDaysAgo, twoDaysAgo.Add(1*time.Hour), nil)
	assert.NoError(t, err)
	_, err = db.CreateManualEntry("project", "recent", yesterday, yesterday.Add(1*time.Hour), nil)
	assert.NoError(t, err)
	_, err = db.CreateManualEntry("project", "today", now.Add(-1*time.Hour), now, nil)
	assert.NoError(t, err)

	// Get entries from yesterday onwards
	entries, err := db.GetEntriesByDateRange(yesterday.Add(-1*time.Hour), now.Add(1*time.Hour))
	assert.NoError(t, err)
	assert.Len(t, entries, 2) // Should get "recent" and "today"
}

func TestGetAllProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// No projects initially
	projects, err := db.GetAllProjects()
	assert.NoError(t, err)
	assert.Len(t, projects, 0)

	// Create entries for different projects
	_, err = db.CreateEntry("zebra-project", "", nil)
	assert.NoError(t, err)
	_, err = db.CreateEntry("alpha-project", "", nil)
	assert.NoError(t, err)
	_, err = db.CreateEntry("zebra-project", "", nil) // Duplicate
	assert.NoError(t, err)

	// Get all projects
	projects, err = db.GetAllProjects()
	assert.NoError(t, err)
	assert.Len(t, projects, 2)

	// Should be sorted alphabetically
	assert.Equal(t, "alpha-project", projects[0])
	assert.Equal(t, "zebra-project", projects[1])
}

func TestGetProjectsWithCompletedEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create running entry
	_, err := db.CreateEntry("running-project", "", nil)
	assert.NoError(t, err)

	// Create completed entry
	entry, err := db.CreateEntry("completed-project", "", nil)
	assert.NoError(t, err)
	err = db.StopEntry(entry.ID)
	assert.NoError(t, err)

	// Should only return projects with completed entries
	projects, err := db.GetProjectsWithCompletedEntries()
	assert.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "completed-project", projects[0])
}

func TestGetCompletedEntriesByProject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create running entry
	_, err := db.CreateEntry("test-project", "running", nil)
	assert.NoError(t, err)

	// Create completed entries
	entry1, err := db.CreateEntry("test-project", "completed 1", nil)
	assert.NoError(t, err)
	err = db.StopEntry(entry1.ID)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	entry2, err := db.CreateEntry("test-project", "completed 2", nil)
	assert.NoError(t, err)
	err = db.StopEntry(entry2.ID)
	assert.NoError(t, err)

	// Get completed entries
	entries, err := db.GetCompletedEntriesByProject("test-project")
	assert.NoError(t, err)
	assert.Len(t, entries, 2) // Should not include running entry

	// All should have end times
	for _, entry := range entries {
		assert.NotNil(t, entry.EndTime)
	}
}

func TestUpdateTimeEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create an entry
	rate := 100.0
	entry, err := db.CreateEntry("original-project", "original description", &rate)
	assert.NoError(t, err)

	// Update the entry
	newStartTime := time.Now().Add(-2 * time.Hour)
	newEndTime := time.Now().Add(-1 * time.Hour)
	newRate := 150.0

	entry.ProjectName = "updated-project"
	entry.Description = "updated description"
	entry.StartTime = newStartTime
	entry.EndTime = &newEndTime
	entry.HourlyRate = &newRate

	err = db.UpdateTimeEntry(entry.ID, entry)
	assert.NoError(t, err)

	// Verify update
	updated, err := db.GetEntry(entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-project", updated.ProjectName)
	assert.Equal(t, "updated description", updated.Description)
	assert.WithinDuration(t, newStartTime, updated.StartTime, time.Second)
	assert.NotNil(t, updated.EndTime)
	assert.WithinDuration(t, newEndTime, *updated.EndTime, time.Second)
	assert.Equal(t, newRate, *updated.HourlyRate)
}

func TestDeleteTimeEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create an entry
	entry, err := db.CreateEntry("test-project", "to be deleted", nil)
	assert.NoError(t, err)

	// Delete it
	err = db.DeleteTimeEntry(entry.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = db.GetEntry(entry.ID)
	assert.Error(t, err)

	// Delete non-existent entry (should not error)
	err = db.DeleteTimeEntry(9999)
	assert.NoError(t, err)
}

func TestTimeEntryDuration(t *testing.T) {
	entry := &TimeEntry{
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   timePtr(time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC)),
	}

	duration := entry.Duration()
	assert.Equal(t, 90*time.Minute, duration)
}

func TestTimeEntryIsRunning(t *testing.T) {
	tests := []struct {
		name     string
		entry    *TimeEntry
		expected bool
	}{
		{
			name: "running entry",
			entry: &TimeEntry{
				StartTime: time.Now(),
				EndTime:   nil,
			},
			expected: true,
		},
		{
			name: "stopped entry",
			entry: &TimeEntry{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   timePtr(time.Now()),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.entry.IsRunning())
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}
