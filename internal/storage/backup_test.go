package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DylanDevelops/tmpo/internal/fsperm"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// makeSettingsDB creates a SQLite file with a settings table at the given path.
func makeSettingsDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	assert.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	assert.NoError(t, err)
	return db
}

func TestCurrentSchemaVersion(t *testing.T) {
	t.Run("matches length of allMigrationKeys", func(t *testing.T) {
		assert.Equal(t, len(allMigrationKeys), CurrentSchemaVersion)
	})

	t.Run("is greater than zero", func(t *testing.T) {
		assert.Greater(t, CurrentSchemaVersion, 0)
	})
}

func TestGetDBPath(t *testing.T) {
	t.Run("default path ends with tmpo.db under .tmpo", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "")
		path, err := GetDBPath()
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(path, "tmpo.db"))
		assert.Contains(t, path, ".tmpo")
	})

	t.Run("dev mode uses .tmpo-dev directory", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "1")
		path, err := GetDBPath()
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(path, "tmpo.db"))
		assert.Contains(t, path, ".tmpo-dev")
	})
}

func TestGetBackupDir(t *testing.T) {
	t.Run("default backup dir ends with backups under .tmpo", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "")
		dir, err := GetBackupDir()
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(dir, "backups"))
		assert.Contains(t, dir, ".tmpo")
	})

	t.Run("dev mode backup dir is under .tmpo-dev", func(t *testing.T) {
		t.Setenv("TMPO_DEV", "1")
		dir, err := GetBackupDir()
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(dir, "backups"))
		assert.Contains(t, dir, ".tmpo-dev")
	})
}

func TestGetBackupSchemaVersion(t *testing.T) {
	t.Run("returns zero when settings table is absent", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "no_settings.db")

		db, err := sql.Open("sqlite", path)
		assert.NoError(t, err)
		_, err = db.Exec(`CREATE TABLE time_entries (id INTEGER PRIMARY KEY)`)
		assert.NoError(t, err)
		db.Close()

		version, err := getBackupSchemaVersion(path)
		assert.NoError(t, err)
		assert.Equal(t, 0, version)
	})

	t.Run("returns zero when settings table is empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty_settings.db")

		db := makeSettingsDB(t, path)
		db.Close()

		version, err := getBackupSchemaVersion(path)
		assert.NoError(t, err)
		assert.Equal(t, 0, version)
	})

	t.Run("counts completed migrations correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "migrated.db")

		db := makeSettingsDB(t, path)
		_, err := db.Exec(
			"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
			Migration001_UTCTimestamps, "completed", time.Now().UTC(),
		)
		assert.NoError(t, err)
		db.Close()

		version, err := getBackupSchemaVersion(path)
		assert.NoError(t, err)
		assert.Equal(t, 1, version)
	})

	t.Run("does not count non-completed migration entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "pending.db")

		db := makeSettingsDB(t, path)
		_, err := db.Exec(
			"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
			Migration001_UTCTimestamps, "pending", time.Now().UTC(),
		)
		assert.NoError(t, err)
		db.Close()

		version, err := getBackupSchemaVersion(path)
		assert.NoError(t, err)
		assert.Equal(t, 0, version)
	})
}

func TestListBackups_EmptyWhenDirAbsent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	backups, err := ListBackups()
	assert.NoError(t, err)
	assert.Empty(t, backups)
}

func TestListBackups_IgnoresNonDBFiles(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	backupDir := filepath.Join(tmpHome, ".tmpo", "backups")
	assert.NoError(t, os.MkdirAll(backupDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(backupDir, "notes.txt"), []byte("ignore"), 0644))

	backups, err := ListBackups()
	assert.NoError(t, err)
	assert.Empty(t, backups)
}

func TestListBackups_SortedOldestFirst(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	backupDir := filepath.Join(tmpHome, ".tmpo", "backups")
	assert.NoError(t, os.MkdirAll(backupDir, 0755))

	olderPath := filepath.Join(backupDir, "tmpo-20260101-100000.db")
	newerPath := filepath.Join(backupDir, "tmpo-20260102-100000.db")

	for _, path := range []string{olderPath, newerPath} {
		db := makeSettingsDB(t, path)
		db.Close()
	}

	now := time.Now()
	assert.NoError(t, os.Chtimes(olderPath, now.Add(-time.Hour), now.Add(-time.Hour)))
	assert.NoError(t, os.Chtimes(newerPath, now, now))

	backups, err := ListBackups()
	assert.NoError(t, err)
	assert.Len(t, backups, 2)

	assert.Equal(t, 1, backups[0].ID)
	assert.Equal(t, 2, backups[1].ID)
	assert.Equal(t, "tmpo-20260101-100000.db", backups[0].Filename)
	assert.Equal(t, "tmpo-20260102-100000.db", backups[1].Filename)
	assert.True(t, backups[0].CreatedAt.Before(backups[1].CreatedAt))
}

func TestCreateBackup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	db, err := Initialize()
	assert.NoError(t, err)
	defer db.Close()

	backup, err := db.CreateBackup()
	assert.NoError(t, err)
	assert.NotNil(t, backup)

	assert.True(t, strings.HasPrefix(backup.Filename, "tmpo-"))
	assert.True(t, strings.HasSuffix(backup.Filename, ".db"))
	assert.Greater(t, backup.Size, int64(0))
	assert.Equal(t, CurrentSchemaVersion, backup.SchemaVersion)
	assert.True(t, backup.IsUpToDate)

	_, err = os.Stat(backup.Path)
	assert.NoError(t, err)
}

func TestRestoreBackup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	// Create and populate the live DB, then snapshot it
	db, err := Initialize()
	assert.NoError(t, err)
	_, err = db.CreateEntry("test-project", "before backup", nil, nil)
	assert.NoError(t, err)

	backup, err := db.CreateBackup()
	assert.NoError(t, err)
	db.Close()

	// Add an entry after the backup was taken
	db2, err := Initialize()
	assert.NoError(t, err)
	_, err = db2.CreateEntry("test-project", "after backup", nil, nil)
	assert.NoError(t, err)
	entries, err := db2.GetEntries(0)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	db2.Close()

	// Restore and verify only the pre-backup state remains
	assert.NoError(t, RestoreBackup(backup.Path))

	db3, err := Initialize()
	assert.NoError(t, err)
	defer db3.Close()

	entries, err = db3.GetEntries(0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "before backup", entries[0].Description)
}

func TestCreateBackup_SetsPrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	db, err := Initialize()
	assert.NoError(t, err)
	defer db.Close()

	info, err := db.CreateBackup()
	assert.NoError(t, err)

	dirInfo, err := os.Stat(filepath.Join(tmpHome, ".tmpo", "backups"))
	assert.NoError(t, err)
	assert.Equal(t, fsperm.DirPerm, dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(info.Path)
	assert.NoError(t, err)
	assert.Equal(t, fsperm.FilePerm, fileInfo.Mode().Perm())
}

func TestRestoreBackup_SetsPrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	db, err := Initialize()
	assert.NoError(t, err)
	info, err := db.CreateBackup()
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	assert.NoError(t, RestoreBackup(info.Path))

	dbInfo, err := os.Stat(filepath.Join(tmpHome, ".tmpo", "tmpo.db"))
	assert.NoError(t, err)
	assert.Equal(t, fsperm.FilePerm, dbInfo.Mode().Perm())
}
