package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUndoTestDB(t *testing.T) *Database {
	t.Helper()
	db := setupTestDB(t)
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	require.NoError(t, err)
	return db
}

func TestSaveAndGetLastAction(t *testing.T) {
	db := setupUndoTestDB(t)
	defer db.Close()

	action := UndoAction{Type: ActionStop, EntryID: 42, ProjectName: "myproject"}
	require.NoError(t, db.SaveLastAction(action))

	got, err := db.GetLastAction()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, ActionStop, got.Type)
	assert.Equal(t, int64(42), got.EntryID)
	assert.Equal(t, "myproject", got.ProjectName)
}

func TestGetLastAction_WhenNone(t *testing.T) {
	db := setupUndoTestDB(t)
	defer db.Close()

	got, err := db.GetLastAction()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSaveLastAction_Overwrites(t *testing.T) {
	db := setupUndoTestDB(t)
	defer db.Close()

	require.NoError(t, db.SaveLastAction(UndoAction{Type: ActionStop, EntryID: 1, ProjectName: "first"}))
	require.NoError(t, db.SaveLastAction(UndoAction{Type: ActionStart, EntryID: 2, ProjectName: "second"}))

	got, err := db.GetLastAction()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, ActionStart, got.Type)
	assert.Equal(t, int64(2), got.EntryID)
}

func TestClearLastAction(t *testing.T) {
	db := setupUndoTestDB(t)
	defer db.Close()

	require.NoError(t, db.SaveLastAction(UndoAction{Type: ActionStop, EntryID: 1, ProjectName: "proj"}))
	require.NoError(t, db.ClearLastAction())

	got, err := db.GetLastAction()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSaveLastAction_PreservesEntrySnapshot(t *testing.T) {
	db := setupUndoTestDB(t)
	defer db.Close()

	rate := 75.0
	milestone := "v1"
	end := time.Now()
	entry := &TimeEntry{
		ID:            99,
		ProjectName:   "proj",
		StartTime:     time.Now().Add(-time.Hour),
		EndTime:       &end,
		Description:   "some work",
		HourlyRate:    &rate,
		MilestoneName: &milestone,
	}

	require.NoError(t, db.SaveLastAction(UndoAction{Type: ActionDelete, ProjectName: "proj", Entry: entry}))

	got, err := db.GetLastAction()
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Entry)
	assert.Equal(t, int64(99), got.Entry.ID)
	assert.Equal(t, "some work", got.Entry.Description)
	assert.Equal(t, 75.0, *got.Entry.HourlyRate)
	assert.Equal(t, "v1", *got.Entry.MilestoneName)
}

func TestUncompleteEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	entry, err := db.CreateEntry("proj", "", nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.StopEntry(entry.ID))

	stopped, err := db.GetEntry(entry.ID)
	require.NoError(t, err)
	assert.NotNil(t, stopped.EndTime)

	require.NoError(t, db.UncompleteEntry(entry.ID))

	running, err := db.GetEntry(entry.ID)
	require.NoError(t, err)
	assert.Nil(t, running.EndTime)
}

func TestRestoreDeletedEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	start := time.Now().Add(-2 * time.Hour)
	end := time.Now().Add(-time.Hour)
	rate := 100.0
	milestone := "m1"
	original, err := db.CreateManualEntry("proj", "work", start, end, &rate, &milestone)
	require.NoError(t, err)

	require.NoError(t, db.DeleteTimeEntry(original.ID))

	require.NoError(t, db.RestoreDeletedEntry(original))

	restored, err := db.GetEntry(original.ID)
	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.ProjectName, restored.ProjectName)
	assert.Equal(t, original.Description, restored.Description)
	assert.Equal(t, *original.HourlyRate, *restored.HourlyRate)
	assert.Equal(t, *original.MilestoneName, *restored.MilestoneName)
	assert.NotNil(t, restored.EndTime)
}

func TestRestoreDeletedEntry_RunningEntry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	original, err := db.CreateEntry("proj", "active", nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.DeleteTimeEntry(original.ID))

	require.NoError(t, db.RestoreDeletedEntry(original))

	restored, err := db.GetEntry(original.ID)
	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Nil(t, restored.EndTime)
}
