package utilities

import (
	"testing"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUndoTestDB(t *testing.T) *storage.Database {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "1")
	db, err := storage.Initialize()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUndoActionDescription(t *testing.T) {
	tests := []struct {
		actionType storage.ActionType
		contains   string
	}{
		{storage.ActionStop, "Stopped tracking"},
		{storage.ActionPause, "Paused tracking"},
		{storage.ActionStart, "Started tracking"},
		{storage.ActionResume, "Resumed tracking"},
		{storage.ActionManual, "Created manual entry for"},
		{storage.ActionDelete, "Deleted entry for"},
	}

	for _, tt := range tests {
		t.Run(string(tt.actionType), func(t *testing.T) {
			action := &storage.UndoAction{Type: tt.actionType, ProjectName: "proj"}
			desc := undoActionDescription(action)
			assert.Contains(t, desc, tt.contains)
			assert.Contains(t, desc, "proj")
		})
	}
}

func TestUndoActionDescription_Unknown(t *testing.T) {
	action := &storage.UndoAction{Type: "something_new", ProjectName: "proj"}
	desc := undoActionDescription(action)
	assert.Contains(t, desc, "Unknown action")
	assert.Contains(t, desc, "something_new")
}

func TestApplyUndo_Stop_ErrorWhenTimerAlreadyRunning(t *testing.T) {
	db := setupUndoTestDB(t)

	stopped, err := db.CreateEntry("proj", "", nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.StopEntry(stopped.ID))

	_, err = db.CreateEntry("other", "", nil, nil)
	require.NoError(t, err)

	action := &storage.UndoAction{Type: storage.ActionStop, EntryID: stopped.ID, ProjectName: "proj"}
	err = applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timer is already running")
}

func TestApplyUndo_Pause_ErrorWhenTimerAlreadyRunning(t *testing.T) {
	db := setupUndoTestDB(t)

	stopped, err := db.CreateEntry("proj", "", nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.StopEntry(stopped.ID))

	_, err = db.CreateEntry("other", "", nil, nil)
	require.NoError(t, err)

	action := &storage.UndoAction{Type: storage.ActionPause, EntryID: stopped.ID, ProjectName: "proj"}
	err = applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timer is already running")
}

func TestApplyUndo_Delete_ErrorWhenNoSnapshot(t *testing.T) {
	db := setupUndoTestDB(t)

	action := &storage.UndoAction{Type: storage.ActionDelete, ProjectName: "proj", Entry: nil}
	err := applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entry snapshot")
}

func TestApplyUndo_UnknownType_ReturnsError(t *testing.T) {
	db := setupUndoTestDB(t)

	action := &storage.UndoAction{Type: "bogus", ProjectName: "proj"}
	err := applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action type")
}
