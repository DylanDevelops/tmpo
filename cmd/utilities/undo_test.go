package utilities

import (
	"testing"
	"time"

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
		{storage.ActionEdit, "Edited entry for"},
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

func TestApplyUndo_Edit_ErrorWhenNoSnapshot(t *testing.T) {
	db := setupUndoTestDB(t)

	action := &storage.UndoAction{Type: storage.ActionEdit, ProjectName: "proj", Entry: nil}
	err := applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entry snapshot")
}

func TestApplyUndo_Edit_RestoresOriginalEntry(t *testing.T) {
	db := setupUndoTestDB(t)

	// Create and stop an entry, then capture its pre-edit snapshot.
	orig, err := db.CreateEntry("proj", "before", nil, nil)
	require.NoError(t, err)
	require.NoError(t, db.StopEntry(orig.ID))

	before, err := db.GetEntry(orig.ID)
	require.NoError(t, err)

	// Apply an edit that changes several fields.
	edited := *before
	newStart := before.StartTime.Add(-time.Hour)
	newEnd := before.EndTime.Add(time.Hour)
	edited.StartTime = newStart
	edited.EndTime = &newEnd
	edited.Description = "after"
	require.NoError(t, db.UpdateTimeEntry(edited.ID, &edited))

	// Undo should restore the original values.
	action := &storage.UndoAction{Type: storage.ActionEdit, EntryID: before.ID, ProjectName: "proj", Entry: before}
	require.NoError(t, applyUndo(db, action))

	restored, err := db.GetEntry(orig.ID)
	require.NoError(t, err)
	assert.Equal(t, "before", restored.Description)
	assert.True(t, restored.StartTime.Equal(before.StartTime), "start time should be restored")
	require.NotNil(t, restored.EndTime)
	assert.True(t, restored.EndTime.Equal(*before.EndTime), "end time should be restored")
}

func TestApplyUndo_UnknownType_ReturnsError(t *testing.T) {
	db := setupUndoTestDB(t)

	action := &storage.UndoAction{Type: "bogus", ProjectName: "proj"}
	err := applyUndo(db, action)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action type")
}
