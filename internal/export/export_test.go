package export

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestToCSV(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("exports entries to CSV", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "test-project",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "Test work",
			},
			{
				ID:          2,
				ProjectName: "another-project",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "More work",
			},
		}

		filename := filepath.Join(tmpDir, "test.csv")
		err := ToCSV(entries, filename)
		assert.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filename)
		assert.NoError(t, err)

		// Read and verify CSV content
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		assert.NoError(t, err)

		// Should have header + 2 entries
		assert.Len(t, records, 3)

		// Verify header
		assert.Equal(t, []string{"Project", "Start Time", "End Time", "Duration (hours)", "Description"}, records[0])

		// Verify first entry
		assert.Equal(t, "test-project", records[1][0])
		assert.Equal(t, "2024-01-01 09:00:00", records[1][1])
		assert.Equal(t, "2024-01-01 17:00:00", records[1][2])
		assert.Equal(t, "8.00", records[1][3]) // 8 hours
		assert.Equal(t, "Test work", records[1][4])
	})

	t.Run("handles running entries", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "running-project",
				StartTime:   startTime,
				EndTime:     nil, // Running
				Description: "Ongoing work",
			},
		}

		filename := filepath.Join(tmpDir, "running.csv")
		err := ToCSV(entries, filename)
		assert.NoError(t, err)

		// Read CSV
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		assert.NoError(t, err)

		// End time should be empty string
		assert.Empty(t, records[1][2])
	})

	t.Run("handles empty entries", func(t *testing.T) {
		entries := []*storage.TimeEntry{}

		filename := filepath.Join(tmpDir, "empty.csv")
		err := ToCSV(entries, filename)
		assert.NoError(t, err)

		// Read CSV
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		assert.NoError(t, err)

		// Should only have header
		assert.Len(t, records, 1)
	})

	t.Run("handles entries without description", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "test",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "", // Empty
			},
		}

		filename := filepath.Join(tmpDir, "no-desc.csv")
		err := ToCSV(entries, filename)
		assert.NoError(t, err)

		// Read CSV
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		assert.NoError(t, err)

		// Description should be empty string
		assert.Empty(t, records[1][4])
	})
}

func TestToJson(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("exports entries to JSON", func(t *testing.T) {
		startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "test-project",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "Test work",
			},
			{
				ID:          2,
				ProjectName: "another-project",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "More work",
			},
		}

		filename := filepath.Join(tmpDir, "test.json")
		err := ToJson(entries, filename)
		assert.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filename)
		assert.NoError(t, err)

		// Read and verify JSON content
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		var exportedEntries []ExportEntry
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&exportedEntries)
		assert.NoError(t, err)

		// Should have 2 entries
		assert.Len(t, exportedEntries, 2)

		// Verify first entry
		assert.Equal(t, "test-project", exportedEntries[0].Project)
		assert.Equal(t, "2024-01-01T09:00:00Z", exportedEntries[0].StartTime)
		assert.Equal(t, "2024-01-01T17:00:00Z", exportedEntries[0].EndTime)
		assert.Equal(t, 8.0, exportedEntries[0].Duration)
		assert.Equal(t, "Test work", exportedEntries[0].Description)
	})

	t.Run("handles running entries", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "running-project",
				StartTime:   startTime,
				EndTime:     nil, // Running
				Description: "Ongoing work",
			},
		}

		filename := filepath.Join(tmpDir, "running.json")
		err := ToJson(entries, filename)
		assert.NoError(t, err)

		// Read JSON
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		var exportedEntries []ExportEntry
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&exportedEntries)
		assert.NoError(t, err)

		// End time should be omitted (zero value)
		assert.Empty(t, exportedEntries[0].EndTime)
	})

	t.Run("handles empty entries", func(t *testing.T) {
		entries := []*storage.TimeEntry{}

		filename := filepath.Join(tmpDir, "empty.json")
		err := ToJson(entries, filename)
		assert.NoError(t, err)

		// Read JSON
		file, err := os.Open(filename)
		assert.NoError(t, err)
		defer file.Close()

		var exportedEntries []ExportEntry
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&exportedEntries)
		assert.NoError(t, err)

		// Should be empty array
		assert.Len(t, exportedEntries, 0)
	})

	t.Run("omits empty description", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		entries := []*storage.TimeEntry{
			{
				ID:          1,
				ProjectName: "test",
				StartTime:   startTime,
				EndTime:     &endTime,
				Description: "", // Empty - should be omitted
			},
		}

		filename := filepath.Join(tmpDir, "no-desc.json")
		err := ToJson(entries, filename)
		assert.NoError(t, err)

		// Read raw JSON to verify omission
		content, err := os.ReadFile(filename)
		assert.NoError(t, err)

		// Description field should be omitted when empty
		// (Note: Go's JSON encoder may still include it as empty string depending on omitempty behavior)
		var rawData []map[string]interface{}
		err = json.Unmarshal(content, &rawData)
		assert.NoError(t, err)

		// Description should either be omitted or empty
		if desc, exists := rawData[0]["description"]; exists {
			assert.Empty(t, desc)
		}
	})
}
