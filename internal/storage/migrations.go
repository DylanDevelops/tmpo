package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// Migration keys
// ! I'm adding this system so that future database migrations will be easier - Dylan
const (
	Migration001_UTCTimestamps = "001_utc_timestamps"
	Migration002_EntrySchema   = "002_entry_schema"
)

// allMigrationKeys lists every migration in order. Adding a new migration here automatically bumps CurrentSchemaVersion.
var allMigrationKeys = []string{
	Migration001_UTCTimestamps,
	Migration002_EntrySchema,
}

// CurrentSchemaVersion is the number of migrations the current binary knows about.
var CurrentSchemaVersion = len(allMigrationKeys)

// runMigrations executes all pending migrations
func (d *Database) runMigrations() error {
	// Migration 1: Convert all timestamps to UTC
	if err := d.migrateTimestampsToUTC(); err != nil {
		return fmt.Errorf("timestamp UTC migration failed: %w", err)
	}

	// Migration 2: Ensure time_entries has the hourly_rate/milestone_name
	// columns and supporting indexes (previously ad-hoc ALTERs in Initialize).
	if err := d.migrateEntrySchema(); err != nil {
		return fmt.Errorf("entry schema migration failed: %w", err)
	}

	return nil
}

// hasPendingMigrations reports whether any known migration has not yet been
// marked complete. Used to decide whether a pre-migration backup is warranted.
func (d *Database) hasPendingMigrations() (bool, error) {
	for _, key := range allMigrationKeys {
		completed, err := d.hasMigrationRun(key)
		if err != nil {
			return false, err
		}
		if !completed {
			return true, nil
		}
	}
	return false, nil
}

// columnExists reports whether the given column is present on the table by
// inspecting PRAGMA table_info. This replaces the old fragile error-string
// matching against "duplicate column name".
func columnExists(tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("failed to read table info for %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("failed to scan table info for %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("error iterating table info for %s: %w", table, err)
	}

	return false, nil
}

// migrateEntrySchema brings older databases up to the current time_entries
// schema by adding any missing columns and indexes.
func (d *Database) migrateEntrySchema() error {
	completed, err := d.hasMigrationRun(Migration002_EntrySchema)
	if err != nil {
		return err
	}

	if completed {
		// migration is already finished
		return nil
	}

	// start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// rollback changes if something explodes
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// add hourly_rate column if it is missing (legacy databases predate it)
	var hasHourlyRate bool
	if hasHourlyRate, err = columnExists(tx, "time_entries", "hourly_rate"); err != nil {
		return err
	}
	if !hasHourlyRate {
		if _, err = tx.Exec(`ALTER TABLE time_entries ADD COLUMN hourly_rate REAL`); err != nil {
			return fmt.Errorf("failed to add hourly_rate column: %w", err)
		}
	}

	// add milestone_name column if it is missing
	var hasMilestoneName bool
	if hasMilestoneName, err = columnExists(tx, "time_entries", "milestone_name"); err != nil {
		return err
	}
	if !hasMilestoneName {
		if _, err = tx.Exec(`ALTER TABLE time_entries ADD COLUMN milestone_name TEXT`); err != nil {
			return fmt.Errorf("failed to add milestone_name column: %w", err)
		}
	}

	// ensure supporting indexes exist
	if _, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_time_entries_milestone ON time_entries(milestone_name)`); err != nil {
		return fmt.Errorf("failed to create milestone index: %w", err)
	}
	if _, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_milestones_project_active ON milestones(project_name, end_time)`); err != nil {
		return fmt.Errorf("failed to create milestones index: %w", err)
	}

	// mark migration as complete in transaction
	_, err = tx.Exec(
		"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
		Migration002_EntrySchema,
		"completed",
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to mark migration complete: %w", err)
	}

	// push changes to db
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

func (d *Database) hasMigrationRun(migrationKey string) (bool, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM settings WHERE key = ?", migrationKey).Scan(&value)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return value == "completed", nil
}

// markMigrationComplete marks a migration as completed
func (d *Database) markMigrationComplete(migrationKey string) error {
	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
		migrationKey,
		"completed",
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to mark migration complete: %w", err)
	}

	return nil
}

func (d *Database) migrateTimestampsToUTC() error {
	completed, err := d.hasMigrationRun(Migration001_UTCTimestamps)
	if err != nil {
		return err
	}

	if completed {
		// migration is already finished
		return nil
	}

	// start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// rollback changes if something explodes
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// migrate time_entries table
	if err = d.migrateTimeEntriesTableToUTC(tx); err != nil {
		return fmt.Errorf("failed to migrate time_entries: %w", err)
	}

	// migrate milestones table
	if err = d.migrateMilestonesTableToUTC(tx); err != nil {
		return fmt.Errorf("failed to migrate milestones: %w", err)
	}

	// mark migration as complete in transaction
	_, err = tx.Exec(
		"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
		Migration001_UTCTimestamps,
		"completed",
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to mark migration complete: %w", err)
	}

	// push changes to db
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

func (d *Database) migrateTimeEntriesTableToUTC(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT id, start_time, end_time FROM time_entries")
	if err != nil {
		return fmt.Errorf("failed to query time_entries: %w", err)
	}
	defer rows.Close()

	type entryUpdate struct {
		id        int64
		startTime time.Time
		endTime   sql.NullTime
	}

	var updates []entryUpdate

	for rows.Next() {
		var entry entryUpdate

		if err := rows.Scan(&entry.id, &entry.startTime, &entry.endTime); err != nil {
			return fmt.Errorf("failed to scan entry: %w", err)
		}

		// check if timestamp needs conversion
		needsUpdate := false

		if entry.startTime.Location() != time.UTC {
			entry.startTime = entry.startTime.UTC()
			needsUpdate = true
		}

		if entry.endTime.Valid && entry.endTime.Time.Location() != time.UTC {
			entry.endTime.Time = entry.endTime.Time.UTC()
			needsUpdate = true
		}

		if needsUpdate {
			updates = append(updates, entry)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating entries: %w", err)
	}

	// apply updates
	for _, update := range updates {
		_, err := tx.Exec(
			"UPDATE time_entries SET start_time = ?, end_time = ? WHERE id = ?",
			update.startTime,
			update.endTime,
			update.id,
		)
		if err != nil {
			return fmt.Errorf("failed to update entry %d: %w", update.id, err)
		}
	}

	return nil
}

func (d *Database) migrateMilestonesTableToUTC(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT id, start_time, end_time FROM milestones")
	if err != nil {
		return fmt.Errorf("failed to query milestones: %w", err)
	}
	defer rows.Close()

	type milestoneUpdate struct {
		id        int64
		startTime time.Time
		endTime   sql.NullTime
	}

	var updates []milestoneUpdate

	for rows.Next() {
		var milestone milestoneUpdate

		if err := rows.Scan(&milestone.id, &milestone.startTime, &milestone.endTime); err != nil {
			return fmt.Errorf("failed to scan milestone: %w", err)
		}

		// check if timestamps is not already UTC
		needsUpdate := false

		if milestone.startTime.Location() != time.UTC {
			milestone.startTime = milestone.startTime.UTC()
			needsUpdate = true
		}

		if milestone.endTime.Valid && milestone.endTime.Time.Location() != time.UTC {
			milestone.endTime.Time = milestone.endTime.Time.UTC()
			needsUpdate = true
		}

		if needsUpdate {
			updates = append(updates, milestone)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating milestones: %w", err)
	}

	// apply updates
	for _, update := range updates {
		_, err := tx.Exec(
			"UPDATE milestones SET start_time = ?, end_time = ? WHERE id = ?",
			update.startTime,
			update.endTime,
			update.id,
		)
		if err != nil {
			return fmt.Errorf("failed to update milestone %d: %w", update.id, err)
		}
	}

	return nil
}
