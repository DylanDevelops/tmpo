package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/DylanDevelops/tmpo/internal/fsperm"
	"github.com/DylanDevelops/tmpo/internal/paths"
	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func getTmpoDir() (string, error) {
	return paths.TmpoDir()
}

func Initialize() (*Database, error) {
	tmpoDir, err := getTmpoDir()
	if err != nil {
		return nil, err
	}

	if err := fsperm.SecureDir(tmpoDir); err != nil {
		return nil, fmt.Errorf("failed to create .tmpo directory: %w", err)
	}

	dbPath := filepath.Join(tmpoDir, "tmpo.db")
	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

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

	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS milestones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			UNIQUE(project_name, name)
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create milestones table: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE time_entries ADD COLUMN hourly_rate REAL`)
	if err != nil && !isColumnExistsError(err) {
		return nil, fmt.Errorf("failed to add hourly_rate column: %w", err)
	}

	_, err = db.Exec(`ALTER TABLE time_entries ADD COLUMN milestone_name TEXT`)
	if err != nil && !isColumnExistsError(err) {
		return nil, fmt.Errorf("failed to add milestone_name column: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_time_entries_milestone ON time_entries(milestone_name)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_milestones_project_active ON milestones(project_name, end_time)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// settings table for tracking migrations and other metadata
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create settings table: %w", err)
	}

	database := &Database{db: db}

	if err := database.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := fsperm.SecureFile(dbPath); err != nil {
		return nil, err
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := fsperm.SecureFile(dbPath + suffix); err != nil {
			return nil, err
		}
	}

	return database, nil
}

func isColumnExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate column name") ||
		strings.Contains(errMsg, "duplicate column")
}

func (d *Database) CreateEntry(projectName, description string, hourlyRate *float64, milestoneName *string) (*TimeEntry, error) {
	var rate sql.NullFloat64
	if hourlyRate != nil {
		rate = sql.NullFloat64{Float64: *hourlyRate, Valid: true}
	}

	var milestone sql.NullString
	if milestoneName != nil {
		milestone = sql.NullString{String: *milestoneName, Valid: true}
	}

	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description, hourly_rate, milestone_name) VALUES (?, ?, ?, ?, ?)",
		projectName,
		time.Now().UTC(),
		description,
		rate,
		milestone,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetEntry(id)
}

func (d *Database) CreateManualEntry(projectName, description string, startTime, endTime time.Time, hourlyRate *float64, milestoneName *string) (*TimeEntry, error) {
	var rate sql.NullFloat64
	if hourlyRate != nil {
		rate = sql.NullFloat64{Float64: *hourlyRate, Valid: true}
	}

	var milestone sql.NullString
	if milestoneName != nil {
		milestone = sql.NullString{String: *milestoneName, Valid: true}
	}

	startTimeUTC := startTime.UTC()
	endTimeUTC := endTime.UTC()

	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, end_time, description, hourly_rate, milestone_name) VALUES (?, ?, ?, ?, ?, ?)",
		projectName,
		startTimeUTC,
		endTimeUTC,
		description,
		rate,
		milestone,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create manual entry: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetEntry(id)
}

func (d *Database) GetRunningEntry() (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get running entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

func (d *Database) GetLastStoppedEntry() (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE end_time IS NOT NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get last stopped entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

func (d *Database) GetLastStoppedEntryByProject(projectName string) (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE end_time IS NOT NULL AND project_name = ?
		ORDER BY start_time DESC
		LIMIT 1
	`, projectName).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get last stopped entry for project: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

func (d *Database) StopEntry(id int64) error {
	_, err := d.db.Exec(
		"UPDATE time_entries SET end_time = ? WHERE id = ?",
		time.Now().UTC(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to stop entry: %w", err)
	}

	return nil
}

func (d *Database) CancelEntry(id int64) error {
	err := d.DeleteTimeEntry(id)

	if err != nil {
		return fmt.Errorf("failed to cancel time entry: %w", err)
	}

	return nil
}

func (d *Database) GetEntry(id int64) (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE id = ?
	`, id).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

func (d *Database) GetEntries(limit int) ([]*TimeEntry, error) {
	query := `
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		ORDER BY start_time DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (d *Database) GetEntriesByProject(projectName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE project_name = ?
		ORDER BY start_time DESC
	`, projectName)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (d *Database) GetEntriesByDateRange(start, end time.Time) ([]*TimeEntry, error) {
	// Convert to UTC for consistent comparison with stored timestamps
	startUTC := start.UTC()
	endUTC := end.UTC()

	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE start_time BETWEEN ? AND ?
		ORDER BY start_time DESC
	`, startUTC, endUTC)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (d *Database) GetAllProjects() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT project_name
		FROM time_entries
		ORDER BY project_name
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	defer rows.Close()

	var projects []string

	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (d *Database) GetProjectsWithCompletedEntries() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT project_name
		FROM time_entries
		WHERE end_time IS NOT NULL
		ORDER BY project_name
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	defer rows.Close()

	var projects []string

	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (d *Database) GetCompletedEntriesByProject(projectName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE project_name = ? AND end_time IS NOT NULL
		ORDER BY start_time DESC
	`, projectName)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (d *Database) UpdateTimeEntry(id int64, entry *TimeEntry) error {
	startTimeUTC := entry.StartTime.UTC()

	var endTime sql.NullTime
	if entry.EndTime != nil {
		endTime = sql.NullTime{Time: entry.EndTime.UTC(), Valid: true}
	}

	var hourlyRate sql.NullFloat64
	if entry.HourlyRate != nil {
		hourlyRate = sql.NullFloat64{Float64: *entry.HourlyRate, Valid: true}
	}

	var milestoneName sql.NullString
	if entry.MilestoneName != nil {
		milestoneName = sql.NullString{String: *entry.MilestoneName, Valid: true}
	}

	_, err := d.db.Exec(`
		UPDATE time_entries
		SET project_name = ?, start_time = ?, end_time = ?, description = ?, hourly_rate = ?, milestone_name = ?
		WHERE id = ?
	`, entry.ProjectName, startTimeUTC, endTime, entry.Description, hourlyRate, milestoneName, id)

	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

func (d *Database) DeleteTimeEntry(id int64) error {
	_, err := d.db.Exec("DELETE FROM time_entries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

func (d *Database) CreateMilestone(projectName, name string) (*Milestone, error) {
	result, err := d.db.Exec(
		"INSERT INTO milestones (project_name, name, start_time) VALUES (?, ?, ?)",
		projectName,
		name,
		time.Now().UTC(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetMilestone(id)
}

func (d *Database) GetMilestone(id int64) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE id = ?",
		id,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

func (d *Database) GetActiveMilestoneForProject(projectName string) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? AND end_time IS NULL ORDER BY start_time DESC LIMIT 1",
		projectName,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get active milestone: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

func (d *Database) GetMilestoneByName(projectName, milestoneName string) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? AND name = ?",
		projectName,
		milestoneName,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get milestone by name: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

func (d *Database) GetMilestonesByProject(projectName string) ([]*Milestone, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? ORDER BY start_time DESC",
		projectName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		var milestone Milestone
		var endTime sql.NullTime

		err := rows.Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}

		if endTime.Valid {
			milestone.EndTime = &endTime.Time
		}

		milestones = append(milestones, &milestone)
	}

	return milestones, nil
}

func (d *Database) GetAllMilestones() ([]*Milestone, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, name, start_time, end_time FROM milestones ORDER BY start_time DESC",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get all milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		var milestone Milestone
		var endTime sql.NullTime

		err := rows.Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}

		if endTime.Valid {
			milestone.EndTime = &endTime.Time
		}

		milestones = append(milestones, &milestone)
	}

	return milestones, nil
}

func (d *Database) FinishMilestone(id int64) error {
	_, err := d.db.Exec(
		"UPDATE milestones SET end_time = ? WHERE id = ?",
		time.Now().UTC(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to finish milestone: %w", err)
	}

	return nil
}

func (d *Database) GetEntriesByMilestone(projectName, milestoneName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name FROM time_entries WHERE project_name = ? AND milestone_name = ? ORDER BY start_time DESC",
		projectName,
		milestoneName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get entries by milestone: %w", err)
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
