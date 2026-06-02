package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
)

func ToCSV(entries []*storage.TimeEntry, filename string, inUtc bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	defer writer.Flush()

	header := []string{"Project", "Start Time", "End Time", "Duration (hours)", "Description", "Milestone"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, entry := range slices.Backward(entries) {
		endTime := ""
		if entry.EndTime != nil {
			endTime = toCorrectCsvTimestamp(*entry.EndTime, inUtc)
		}

		milestoneName := ""
		if entry.MilestoneName != nil {
			milestoneName = *entry.MilestoneName
		}

		duration := entry.Duration().Hours()

		record := []string{
			entry.ProjectName,
			toCorrectCsvTimestamp(entry.StartTime, inUtc),
			endTime,
			fmt.Sprintf("%.2f", duration),
			entry.Description,
			milestoneName,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func toCorrectCsvTimestamp(timestamp time.Time, inUtc bool) string {
	formattedTimestamp := ""

	if inUtc {
		formattedTimestamp += timestamp.UTC().Format("2006-01-02 15:04:05")
	} else {
		formattedTimestamp += settings.ToDisplayTime(timestamp).Format("2006-01-02 15:04:05")
	}

	return formattedTimestamp
}
