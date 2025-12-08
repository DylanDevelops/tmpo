package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database is a thin wrapper around an *sql.DB connection pool.
// It centralizes database access for the package and provides a
// single place to implement queries, transactions, migrations and
// other persistence-related helpers.
//
// The underlying *sql.DB is safe for concurrent use, but a Database
// with a nil internal pointer is not usable — it must be initialized
// (for example via sql.Open) and closed when no longer needed.
type Database struct {
	db *sql.DB
}

// Initialize ensures the on-disk storage for the application exists, opens the
// SQLite database, and returns a Database wrapper.
// 
// Specifically, Initialize:
//   - determines the current user's home directory,
//   - creates the directory "$HOME/.tmpo" if it does not already exist,
//   - opens (or creates) the SQLite database file "$HOME/.tmpo/tmpo.db",
//   - ensures the time_entries table exists with the schema:
//       id INTEGER PRIMARY KEY AUTOINCREMENT,
//       project_name TEXT NOT NULL,
//       start_time DATETIME NOT NULL,
//       end_time DATETIME,
//       description TEXT
//
// On success it returns a *Database containing the opened *sql.DB. It returns
// a non-nil error if any step fails: obtaining the home directory, creating
// the directory, opening the database, or creating the table. The caller is
// responsible for closing the database when finished.
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

// CreateEntry inserts a new time entry for the specified projectName with the given
// description. The entry's start_time is set to the current time. On success it returns
// the created *TimeEntry (retrieved by querying the database for the last insert id).
// If the insert or the subsequent retrieval fails, an error wrapping the underlying
// database error is returned.
func (d *Database) CreateEntry(projectName, description string) (*TimeEntry, error) {
	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
		projectName,
		time.Now(),
		description,
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

// GetRunningEntry retrieves the most recently started time entry that is still running
// (i.e. has a NULL end_time) from the time_entries table. The query orders by
// start_time descending and returns at most one row.
//
// If there is no running entry, GetRunningEntry returns (nil, nil). If the database
// query or scan fails, it returns a non-nil error describing the failure.
//
// The function scans id, project_name, start_time, end_time and description into a
// TimeEntry. The EndTime field on the returned TimeEntry is set only if the scanned
// end_time is non-NULL (sql.NullTime.Valid).
func (d *Database) GetRunningEntry() (*TimeEntry, error) {
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
		return nil, fmt.Errorf("failed to get running entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	return &entry, nil
}

// StopEntry sets the end_time of the time entry identified by id to the current time.
// It updates the corresponding row in the time_entries table using time.Now().
// If the update fails (for example if the row does not exist or the database returns an error),
// an error is returned wrapped with context. This method overwrites any existing end_time value
// and does not return the updated entry or perform additional validation on id.
func (d *Database) StopEntry(id int64) error {
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

// GetEntry retrieves a TimeEntry by its ID from the database.
// It queries the time_entries table for id, project_name, start_time, end_time and description,
// scans the result into a TimeEntry value, and returns a pointer to it.
// If the end_time column is NULL in the database, the returned TimeEntry.EndTime will be nil;
// otherwise EndTime will point to the retrieved time value.
// If no row is found or an error occurs during query/scan, an error is returned (wrapped with
// the context "failed to get entry").
func (d *Database) GetEntry(id int64) (*TimeEntry, error) {
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

// GetEntries retrieves time entries from the Database.
//
// It returns time entries ordered by start_time in descending order. If limit > 0,
// at most `limit` entries are returned; if limit <= 0 all matching entries are returned.
// Each returned element is a pointer to a TimeEntry. The EndTime field of a TimeEntry
// will be nil when the corresponding end_time column in the database is NULL.
//
// The function performs a SQL query selecting id, project_name, start_time, end_time,
// and description. It returns a slice of entries and an error if the query or row
// scanning fails; any underlying error is wrapped.
func (d *Database) GetEntries(limit int) ([]*TimeEntry, error) {
	query := `
		SELECT id, project_name, start_time, end_time, description
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

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetEntriesByProject retrieves time entries for the specified projectName from the
// time_entries table. Results are ordered by start_time in descending order (newest first).
//
// For each row a TimeEntry is populated. If the end_time column is NULL the returned
// TimeEntry.EndTime will be nil; otherwise EndTime will point to the scanned time.Time.
//
// On success the function returns a slice of pointers to TimeEntry. If there are no
// matching rows the returned slice will have length 0 (it may be nil). On failure the
// function returns a non-nil error and a nil slice. Errors may originate from the
// query execution, row scanning, or row iteration.
func (d *Database) GetEntriesByProject(projectName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description
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

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetEntriesByDateRange retrieves time entries whose start_time falls between start and end (inclusive).
// Results are returned in descending order by start_time.
// The provided start and end times are passed to the database driver as-is; callers should ensure they use the intended timezone/representation.
// For rows with a NULL end_time the returned TimeEntry.EndTime will be nil; otherwise EndTime points to the parsed time value.
// Returns a slice of pointers to TimeEntry (which may be empty) or an error if the database query or row scanning fails.
func (d *Database) GetEntriesByDateRange(start, end time.Time) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description
		FROM time_entries
		WHERE start_time BETWEEN ? AND ?
		ORDER BY start_time DESC
	`, start, end)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()
	
	var entries []*TimeEntry
	
	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetAllProjects retrieves all distinct project names from the time_entries table.
// The results are returned in ascending order by project_name.
// On success it returns a slice of project names (which will be empty if no projects exist)
// and a nil error. If the underlying database query or a row scan fails, it returns a
// non-nil error describing the failure.
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

// Close closes the Database, releasing any underlying resources.
// It delegates to the wrapped database's Close method and returns any error encountered.
// After Close is called, the Database must not be used for further operations.
func (d *Database) Close() error {
	return d.db.Close()
}