package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Database struct {
	db* sql.DB
}

func Initialize() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	tmpoDir := filepath.Join(homeDir, ".tmpo")
	if err := os.MkdirAll(tmpoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .tmpo directory: %w", err)
	}

	dbPath := filepath.Join(tmpoDir, "tmpo.db")
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			description TEXT
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Database{db: db}, nil
}

func (d* Database) CreateEntry(projectName, description string) (*TimeEntry, error) {
	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
		projectName,
		time.Now(),
		description,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to create entry: %w", err)
	}
	
	id, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("Failed to get last insert id: %w", err)
	}

	return d.GetEntry(id)
}

func (d* Database) GetRunningEntry() (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description
		FROM time_entries
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to get running entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	return &entry, nil
}

func (d* Database) StopEntry(id int64) error {
	_, err := d.db.Exec(
		"UPDATE time_entries SET end_time = ? WHERE id = ?",
		time.Now(),
		id,
	)

	if(err != nil) {
		return fmt.Errorf("failed to stop entry: %w", err)
	}

	return nil
}

func (d* Database) GetEntry(id int64) (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description
		FROM time_entries
		WHERE id = ?
	`, id).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	return &entry, nil
}

func (d* Database) Close() error {
	return d.db.Close()
}