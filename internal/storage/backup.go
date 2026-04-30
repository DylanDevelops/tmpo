package storage

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type BackupInfo struct {
	ID            int
	Filename      string
	Path          string
	CreatedAt     time.Time
	Size          int64
	SchemaVersion int
	IsUpToDate    bool
}

func GetDBPath() (string, error) {
	tmpoDir, err := getTmpoDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tmpoDir, "tmpo.db"), nil
}

func GetBackupDir() (string, error) {
	tmpoDir, err := getTmpoDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tmpoDir, "backups"), nil
}

// CreateBackup uses SQLite's VACUUM INTO to produce a clean, consistent snapshot of the live database.
func (d *Database) CreateBackup() (*BackupInfo, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backups directory: %w", err)
	}

	now := time.Now()
	filename := fmt.Sprintf("tmpo-%s.db", now.Format("20060102-150405"))
	destPath := filepath.Join(backupDir, filename)

	escapedPath := strings.ReplaceAll(destPath, "'", "''")
	if _, err = d.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", escapedPath)); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	info, err := os.Stat(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	return &BackupInfo{
		ID:            1,
		Filename:      filename,
		Path:          destPath,
		CreatedAt:     now,
		Size:          info.Size(),
		SchemaVersion: CurrentSchemaVersion,
		IsUpToDate:    true,
	}, nil
}

// ListBackups returns all backups sorted newest-first with display IDs assigned (1 = newest).
func ListBackups() ([]BackupInfo, error) {
	backupDir, err := GetBackupDir()
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(backupDir)
	if os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read backups directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range dirEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(backupDir, entry.Name())
		version, err := getBackupSchemaVersion(path)
		if err != nil {
			version = 0
		}

		backups = append(backups, BackupInfo{
			Filename:      entry.Name(),
			Path:          path,
			CreatedAt:     info.ModTime(),
			Size:          info.Size(),
			SchemaVersion: version,
			IsUpToDate:    version == CurrentSchemaVersion,
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	for i := range backups {
		backups[i].ID = i + 1
	}

	return backups, nil
}

// RestoreBackup copies a backup file over the live database. The caller must ensure no DB connection is open.
func RestoreBackup(backupPath string) error {
	dbPath, err := GetDBPath()
	if err != nil {
		return err
	}

	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database for writing: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// getBackupSchemaVersion opens a backup SQLite file and counts how many known migrations are marked complete.
func getBackupSchemaVersion(dbPath string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	count := 0
	for _, key := range allMigrationKeys {
		var value string
		err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
		if err == nil && value == "completed" {
			count++
		}
	}

	return count, nil
}
