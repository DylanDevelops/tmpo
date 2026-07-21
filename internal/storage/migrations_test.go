package storage

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupMigrationTestDB creates an in-memory database with settings table for migration testing
func setupMigrationTestDB(t *testing.T) *Database {
	db, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	// Create settings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	assert.NoError(t, err)

	// Create time_entries table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			description TEXT,
			hourly_rate REAL,
			milestone_name TEXT
		)
	`)
	assert.NoError(t, err)

	// Create milestones table
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
	assert.NoError(t, err)

	return &Database{db: db}
}

func TestHasMigrationRun(t *testing.T) {
	tests := []struct {
		name           string
		migrationKey   string
		setupMigration bool
		expected       bool
		expectError    bool
	}{
		{
			name:           "migration not run",
			migrationKey:   "001_test_migration",
			setupMigration: false,
			expected:       false,
			expectError:    false,
		},
		{
			name:           "migration completed",
			migrationKey:   "001_test_migration",
			setupMigration: true,
			expected:       true,
			expectError:    false,
		},
		{
			name:           "different migration key",
			migrationKey:   "002_another_migration",
			setupMigration: false,
			expected:       false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupMigrationTestDB(t)
			defer db.Close()

			if tt.setupMigration {
				_, err := db.db.Exec(
					"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
					tt.migrationKey,
					"completed",
					time.Now().UTC(),
				)
				assert.NoError(t, err)
			}

			hasRun, err := db.hasMigrationRun(tt.migrationKey)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, hasRun)
			}
		})
	}
}

func TestMarkMigrationComplete(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	migrationKey := "001_test_migration"

	// Mark migration as complete
	err := db.markMigrationComplete(migrationKey)
	assert.NoError(t, err)

	// Verify it was marked
	var value string
	var updatedAt time.Time
	err = db.db.QueryRow(
		"SELECT value, updated_at FROM settings WHERE key = ?",
		migrationKey,
	).Scan(&value, &updatedAt)
	assert.NoError(t, err)
	assert.Equal(t, "completed", value)
	assert.WithinDuration(t, time.Now().UTC(), updatedAt, 2*time.Second)

	// Verify hasMigrationRun returns true
	hasRun, err := db.hasMigrationRun(migrationKey)
	assert.NoError(t, err)
	assert.True(t, hasRun)
}

func TestMigrateTimestampsToUTC_FreshDatabase(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Run migration on fresh database (no entries)
	err := db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify migration was marked as complete
	hasRun, err := db.hasMigrationRun(Migration001_UTCTimestamps)
	assert.NoError(t, err)
	assert.True(t, hasRun)
}

func TestMigrateTimestampsToUTC_LocalToUTC(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entries with local timezone
	est, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)

	localTime := time.Date(2026, 1, 8, 15, 30, 0, 0, est) // 3:30 PM EST
	expectedUTC := localTime.UTC()                        // Should convert to 8:30 PM UTC

	// Insert time entry with local timezone
	_, err = db.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, end_time, description) VALUES (?, ?, ?, ?)",
		"test-project",
		localTime,
		localTime.Add(1*time.Hour),
		"Test entry",
	)
	assert.NoError(t, err)

	// Insert milestone with local timezone
	_, err = db.db.Exec(
		"INSERT INTO milestones (project_name, name, start_time, end_time) VALUES (?, ?, ?, ?)",
		"test-project",
		"Sprint 1",
		localTime,
		localTime.Add(7*24*time.Hour),
	)
	assert.NoError(t, err)

	// Run migration
	err = db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify time_entry was converted to UTC
	var entryStartTime, entryEndTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time, end_time FROM time_entries WHERE id = 1",
	).Scan(&entryStartTime, &entryEndTime)
	assert.NoError(t, err)
	assert.Equal(t, time.UTC, entryStartTime.Location())
	assert.Equal(t, time.UTC, entryEndTime.Location())
	assert.Equal(t, expectedUTC.Hour(), entryStartTime.Hour())
	assert.Equal(t, expectedUTC.Minute(), entryStartTime.Minute())

	// Verify milestone was converted to UTC
	var milestoneStartTime, milestoneEndTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time, end_time FROM milestones WHERE id = 1",
	).Scan(&milestoneStartTime, &milestoneEndTime)
	assert.NoError(t, err)
	assert.Equal(t, time.UTC, milestoneStartTime.Location())
	assert.Equal(t, time.UTC, milestoneEndTime.Location())
}

func TestMigrateTimestampsToUTC_Idempotent(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entry with local timezone
	est, err := time.LoadLocation("America/New_York")
	assert.NoError(t, err)

	localTime := time.Date(2026, 1, 8, 15, 30, 0, 0, est)

	_, err = db.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
		"test-project",
		localTime,
		"Test entry",
	)
	assert.NoError(t, err)

	// Run migration first time
	err = db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Get the converted time
	var firstRunTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time FROM time_entries WHERE id = 1",
	).Scan(&firstRunTime)
	assert.NoError(t, err)

	// Run migration second time (should be idempotent)
	err = db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Get the time after second run
	var secondRunTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time FROM time_entries WHERE id = 1",
	).Scan(&secondRunTime)
	assert.NoError(t, err)

	// Times should be identical (migration should skip since already completed)
	assert.Equal(t, firstRunTime, secondRunTime)
}

func TestMigrateTimestampsToUTC_AlreadyUTC(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entry already in UTC
	utcTime := time.Date(2026, 1, 8, 20, 30, 0, 0, time.UTC)

	_, err := db.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
		"test-project",
		utcTime,
		"Already UTC entry",
	)
	assert.NoError(t, err)

	// Run migration
	err = db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify time is unchanged
	var startTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time FROM time_entries WHERE id = 1",
	).Scan(&startTime)
	assert.NoError(t, err)
	assert.Equal(t, utcTime, startTime)
	assert.Equal(t, time.UTC, startTime.Location())
}

func TestMigrateTimestampsToUTC_MixedTimezones(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entries with different timezones
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")

	entries := []struct {
		projectName string
		startTime   time.Time
	}{
		{"project1", time.Date(2026, 1, 8, 15, 0, 0, 0, est)},
		{"project2", time.Date(2026, 1, 8, 12, 0, 0, 0, pst)},
		{"project3", time.Date(2026, 1, 8, 20, 0, 0, 0, time.UTC)},
	}

	for _, entry := range entries {
		_, err := db.db.Exec(
			"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
			entry.projectName,
			entry.startTime,
			"Test",
		)
		assert.NoError(t, err)
	}

	// Run migration
	err := db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify all entries are now UTC
	rows, err := db.db.Query("SELECT start_time FROM time_entries ORDER BY id")
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var startTime time.Time
		err = rows.Scan(&startTime)
		assert.NoError(t, err)
		assert.Equal(t, time.UTC, startTime.Location(), "All timestamps should be in UTC")
	}
}

func TestMigrateTimestampsToUTC_NullEndTimes(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entry with NULL end_time (running entry)
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2026, 1, 8, 15, 30, 0, 0, est)

	_, err := db.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, end_time, description) VALUES (?, ?, NULL, ?)",
		"test-project",
		localTime,
		"Running entry",
	)
	assert.NoError(t, err)

	// Run migration
	err = db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify start_time is UTC and end_time is still NULL
	var startTime time.Time
	var endTime sql.NullTime
	err = db.db.QueryRow(
		"SELECT start_time, end_time FROM time_entries WHERE id = 1",
	).Scan(&startTime, &endTime)
	assert.NoError(t, err)
	assert.Equal(t, time.UTC, startTime.Location())
	assert.False(t, endTime.Valid, "end_time should remain NULL")
}

func TestMigrateTimestampsToUTC_MultipleEntries(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create multiple entries
	est, _ := time.LoadLocation("America/New_York")

	for i := range 10 {
		localTime := time.Date(2026, 1, 8, 15+i, 0, 0, 0, est)
		_, err := db.db.Exec(
			"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
			"test-project",
			localTime,
			"Entry",
		)
		assert.NoError(t, err)
	}

	// Run migration
	err := db.migrateTimestampsToUTC()
	assert.NoError(t, err)

	// Verify all 10 entries are UTC
	var count int
	err = db.db.QueryRow(
		"SELECT COUNT(*) FROM time_entries",
	).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)

	// Verify migration marked as complete
	hasRun, err := db.hasMigrationRun(Migration001_UTCTimestamps)
	assert.NoError(t, err)
	assert.True(t, hasRun)
}

func TestRunMigrations(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Create entry with local timezone
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2026, 1, 8, 15, 30, 0, 0, est)

	_, err := db.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description) VALUES (?, ?, ?)",
		"test-project",
		localTime,
		"Test entry",
	)
	assert.NoError(t, err)

	// Run all migrations
	err = db.runMigrations()
	assert.NoError(t, err)

	// Verify UTC migration ran
	hasRun, err := db.hasMigrationRun(Migration001_UTCTimestamps)
	assert.NoError(t, err)
	assert.True(t, hasRun)

	// Verify entry was converted
	var startTime time.Time
	err = db.db.QueryRow(
		"SELECT start_time FROM time_entries WHERE id = 1",
	).Scan(&startTime)
	assert.NoError(t, err)
	assert.Equal(t, time.UTC, startTime.Location())
}

// setupLegacyEntriesTestDB creates an in-memory database whose time_entries
// table predates the hourly_rate and milestone_name columns, mimicking a very
// old tmpo installation.
func setupLegacyEntriesTestDB(t *testing.T) *Database {
	db, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			description TEXT
		)
	`)
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE milestones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			UNIQUE(project_name, name)
		)
	`)
	assert.NoError(t, err)

	return &Database{db: db}
}

// hasColumn is a test helper that reports whether a table has a given column.
func hasColumn(t *testing.T, db *Database, table, column string) bool {
	t.Helper()
	rows, err := db.db.Query("PRAGMA table_info(" + table + ")")
	assert.NoError(t, err)
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
		assert.NoError(t, rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk))
		if name == column {
			return true
		}
	}
	return false
}

func TestMigrateEntrySchema_AddsMissingColumns(t *testing.T) {
	db := setupLegacyEntriesTestDB(t)
	defer db.Close()

	// Legacy schema lacks both columns
	assert.False(t, hasColumn(t, db, "time_entries", "hourly_rate"))
	assert.False(t, hasColumn(t, db, "time_entries", "milestone_name"))

	err := db.migrateEntrySchema()
	assert.NoError(t, err)

	// Both columns should now exist
	assert.True(t, hasColumn(t, db, "time_entries", "hourly_rate"))
	assert.True(t, hasColumn(t, db, "time_entries", "milestone_name"))

	// Migration should be recorded as complete
	hasRun, err := db.hasMigrationRun(Migration002_EntrySchema)
	assert.NoError(t, err)
	assert.True(t, hasRun)
}

func TestMigrateEntrySchema_Idempotent(t *testing.T) {
	db := setupLegacyEntriesTestDB(t)
	defer db.Close()

	// First run adds the columns
	assert.NoError(t, db.migrateEntrySchema())

	// Second run must be a no-op and must not error on already-present columns
	assert.NoError(t, db.migrateEntrySchema())

	assert.True(t, hasColumn(t, db, "time_entries", "hourly_rate"))
	assert.True(t, hasColumn(t, db, "time_entries", "milestone_name"))
}

func TestMigrateEntrySchema_NoOpWhenColumnsPresent(t *testing.T) {
	// setupMigrationTestDB already includes both columns (current schema)
	db := setupMigrationTestDB(t)
	defer db.Close()

	err := db.migrateEntrySchema()
	assert.NoError(t, err)

	hasRun, err := db.hasMigrationRun(Migration002_EntrySchema)
	assert.NoError(t, err)
	assert.True(t, hasRun)
}

func TestHasPendingMigrations(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	// Nothing marked complete yet
	pending, err := db.hasPendingMigrations()
	assert.NoError(t, err)
	assert.True(t, pending)

	// Run every migration
	assert.NoError(t, db.runMigrations())

	pending, err = db.hasPendingMigrations()
	assert.NoError(t, err)
	assert.False(t, pending)
}

func TestMigrateTimeEntriesTableToUTC_EmptyTable(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	tx, err := db.db.Begin()
	assert.NoError(t, err)
	defer tx.Rollback()

	// Run migration on empty table (should not error)
	err = db.migrateTimeEntriesTableToUTC(tx)
	assert.NoError(t, err)
}

func TestMigrateMilestonesTableToUTC_EmptyTable(t *testing.T) {
	db := setupMigrationTestDB(t)
	defer db.Close()

	tx, err := db.db.Begin()
	assert.NoError(t, err)
	defer tx.Rollback()

	// Run migration on empty table (should not error)
	err = db.migrateMilestonesTableToUTC(tx)
	assert.NoError(t, err)
}
