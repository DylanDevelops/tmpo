package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ActionType string

const (
	ActionStop   ActionType = "stop"
	ActionCancel ActionType = "cancel"
	ActionStart  ActionType = "start"
	ActionPause  ActionType = "pause"
	ActionResume ActionType = "resume"
	ActionDelete ActionType = "delete"
	ActionManual ActionType = "manual"
	ActionEdit   ActionType = "edit"
)

const lastActionKey = "last_action"

type UndoAction struct {
	Type        ActionType `json:"type"`
	EntryID     int64      `json:"entry_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	Entry       *TimeEntry `json:"entry,omitempty"`
}

func (d *Database) SaveLastAction(action UndoAction) error {
	data, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to serialize action: %w", err)
	}
	_, err = d.db.Exec(
		"INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
		lastActionKey,
		string(data),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to save last action: %w", err)
	}
	return nil
}

func (d *Database) GetLastAction() (*UndoAction, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM settings WHERE key = ?", lastActionKey).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last action: %w", err)
	}
	var action UndoAction
	if err := json.Unmarshal([]byte(value), &action); err != nil {
		return nil, fmt.Errorf("failed to parse last action: %w", err)
	}
	return &action, nil
}

func (d *Database) ClearLastAction() error {
	_, err := d.db.Exec("DELETE FROM settings WHERE key = ?", lastActionKey)
	if err != nil {
		return fmt.Errorf("failed to clear last action: %w", err)
	}
	return nil
}

// UncompleteEntry clears the end_time of an entry, resuming it as a running timer.
func (d *Database) UncompleteEntry(id int64) error {
	result, err := d.db.Exec("UPDATE time_entries SET end_time = NULL WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to uncomplete entry: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to uncomplete entry: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("entry %d not found", id)
	}
	return nil
}

// RestoreDeletedEntry re-inserts a previously deleted entry preserving its original ID.
func (d *Database) RestoreDeletedEntry(entry *TimeEntry) error {
	var endTime sql.NullTime
	if entry.EndTime != nil {
		endTime = sql.NullTime{Time: entry.EndTime.UTC(), Valid: true}
	}
	var rate sql.NullFloat64
	if entry.HourlyRate != nil {
		rate = sql.NullFloat64{Float64: *entry.HourlyRate, Valid: true}
	}
	var milestone sql.NullString
	if entry.MilestoneName != nil {
		milestone = sql.NullString{String: *entry.MilestoneName, Valid: true}
	}
	_, err := d.db.Exec(
		"INSERT INTO time_entries (id, project_name, start_time, end_time, description, hourly_rate, milestone_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		entry.ID,
		entry.ProjectName,
		entry.StartTime.UTC(),
		endTime,
		entry.Description,
		rate,
		milestone,
	)
	if err != nil {
		return fmt.Errorf("failed to restore entry: %w", err)
	}
	return nil
}
