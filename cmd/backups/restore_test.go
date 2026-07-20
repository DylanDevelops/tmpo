package backups

import (
	"errors"
	"testing"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupBackupTestHome points the storage layer at an isolated temp home so
// tests never touch the real ~/.tmpo directory.
func setupBackupTestHome(t *testing.T) {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "1")
}

// TestRestoreCmd_BlocksWhenTimerRunning verifies the guard added for issue #134:
// a restore must refuse while a timer is running instead of silently
// obliterating the in-flight entry.
func TestRestoreCmd_BlocksWhenTimerRunning(t *testing.T) {
	setupBackupTestHome(t)

	// Start a running timer, then take a backup so a restore target exists.
	db, err := storage.Initialize()
	require.NoError(t, err)
	_, err = db.CreateEntry("proj", "in flight", nil, nil)
	require.NoError(t, err)
	_, err = db.CreateBackup()
	require.NoError(t, err)
	require.NoError(t, db.Close())

	cmd := RestoreCmd()
	err = cmd.RunE(cmd, []string{})

	// The command must report a handled failure and stop before any restore.
	assert.True(t, errors.Is(err, ui.ErrHandled), "expected ui.ErrHandled, got %v", err)

	// The running entry must survive untouched.
	db2, err := storage.Initialize()
	require.NoError(t, err)
	defer db2.Close()
	running, err := db2.GetRunningEntry()
	require.NoError(t, err)
	require.NotNil(t, running, "running timer should not have been destroyed")
	assert.Equal(t, "in flight", running.Description)
}

// TestRestoreCmd_NoTimerNoBackups confirms the guard passes cleanly when no
// timer is running: with no backups present the command exits without error.
func TestRestoreCmd_NoTimerNoBackups(t *testing.T) {
	setupBackupTestHome(t)

	// Initialize the DB (no running timer, no backups created).
	db, err := storage.Initialize()
	require.NoError(t, err)
	require.NoError(t, db.Close())

	cmd := RestoreCmd()
	err = cmd.RunE(cmd, []string{})

	// No timer and no backups is a clean no-op, not a failure.
	assert.NoError(t, err)
}

// TestRestoreCmd_NoTimerInvalidID confirms that once the guard passes, the
// --id lookup still runs and reports a handled error for an unknown ID.
func TestRestoreCmd_NoTimerInvalidID(t *testing.T) {
	setupBackupTestHome(t)

	// A backup must exist so the command reaches the ID lookup rather than the
	// "no backups found" early return.
	db, err := storage.Initialize()
	require.NoError(t, err)
	_, err = db.CreateBackup()
	require.NoError(t, err)
	require.NoError(t, db.Close())

	cmd := RestoreCmd()
	require.NoError(t, cmd.Flags().Set("id", "9999"))
	err = cmd.RunE(cmd, []string{})

	assert.True(t, errors.Is(err, ui.ErrHandled), "expected ui.ErrHandled, got %v", err)
}
