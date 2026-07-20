package export

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
)

type ExportEntry struct {
	Project     string  `json:"project"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time,omitempty"`
	Duration    float64 `json:"duration_hours"`
	Description string  `json:"description,omitempty"`
	Milestone   string  `json:"milestone,omitempty"`
}

func BuildExportEntries(entries []*storage.TimeEntry, inUtc bool) []ExportEntry {
	exportEntries := make([]ExportEntry, 0, len(entries))

	for _, entry := range slices.Backward(entries) {
		export := ExportEntry{
			Project:     entry.ProjectName,
			StartTime:   toCorrectJsonTimestamp(entry.StartTime, inUtc),
			Duration:    entry.Duration().Hours(),
			Description: entry.Description,
		}

		if entry.EndTime != nil {
			export.EndTime = toCorrectJsonTimestamp(*entry.EndTime, inUtc)
		}

		if entry.MilestoneName != nil {
			export.Milestone = *entry.MilestoneName
		}

		exportEntries = append(exportEntries, export)
	}

	return exportEntries
}

func EncodeJson(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func ToJson(entries []*storage.TimeEntry, filename string, inUtc bool) error {
	exportEntries := BuildExportEntries(entries, inUtc)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}

	defer file.Close()

	return EncodeJson(file, exportEntries)
}

func toCorrectJsonTimestamp(timestamp time.Time, inUtc bool) string {
	formattedTimestamp := ""

	if inUtc {
		formattedTimestamp += timestamp.UTC().Format("2006-01-02T15:04:05Z07:00")
	} else {
		formattedTimestamp += settings.ToDisplayTime(timestamp).Format("2006-01-02T15:04:05Z07:00")
	}

	return formattedTimestamp
}
