package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"github.com/storo/scribescli/pkg/models"
)

// Database manages SQLite storage
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// initSchema creates the database tables
func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS recordings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		audio_path TEXT NOT NULL,
		duration INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		transcript TEXT,
		language TEXT,
		summary TEXT,
		key_points TEXT, -- JSON array
		tags TEXT, -- JSON array
		status TEXT NOT NULL DEFAULT 'completed'
	);

	CREATE TABLE IF NOT EXISTS action_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		recording_id INTEGER NOT NULL,
		priority TEXT NOT NULL,
		task TEXT NOT NULL,
		assignee TEXT,
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_recordings_created_at ON recordings(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_action_items_recording_id ON action_items(recording_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

// SaveRecording saves a recording to the database
func (d *Database) SaveRecording(r *models.Recording) error {
	keyPointsJSON, err := json.Marshal(r.KeyPoints)
	if err != nil {
		return fmt.Errorf("failed to marshal key points: %w", err)
	}

	tagsJSON, err := json.Marshal(r.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now

	result, err := d.db.Exec(`
		INSERT INTO recordings (title, audio_path, duration, created_at, updated_at,
			transcript, language, summary, key_points, tags, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.Title, r.AudioPath, r.Duration, r.CreatedAt, r.UpdatedAt,
		r.Transcript, r.Language, r.Summary, keyPointsJSON, tagsJSON, r.Status)

	if err != nil {
		return fmt.Errorf("failed to insert recording: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	r.ID = id

	// Save action items
	for i := range r.ActionItems {
		r.ActionItems[i].RecordingID = id
		if err := d.SaveActionItem(&r.ActionItems[i]); err != nil {
			return fmt.Errorf("failed to save action item: %w", err)
		}
	}

	return nil
}

// SaveActionItem saves an action item to the database
func (d *Database) SaveActionItem(item *models.ActionItem) error {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	result, err := d.db.Exec(`
		INSERT INTO action_items (recording_id, priority, task, assignee, completed, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, item.RecordingID, item.Priority, item.Task, item.Assignee, item.Completed, item.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert action item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	item.ID = id
	return nil
}

// GetRecording retrieves a recording by ID
func (d *Database) GetRecording(id int64) (*models.Recording, error) {
	var r models.Recording
	var keyPointsJSON, tagsJSON string

	err := d.db.QueryRow(`
		SELECT id, title, audio_path, duration, created_at, updated_at,
			transcript, language, summary, key_points, tags, status
		FROM recordings WHERE id = ?
	`, id).Scan(&r.ID, &r.Title, &r.AudioPath, &r.Duration, &r.CreatedAt, &r.UpdatedAt,
		&r.Transcript, &r.Language, &r.Summary, &keyPointsJSON, &tagsJSON, &r.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("recording not found")
		}
		return nil, fmt.Errorf("failed to query recording: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(keyPointsJSON), &r.KeyPoints); err != nil {
		r.KeyPoints = []string{}
	}
	if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
		r.Tags = []string{}
	}

	// Load action items
	r.ActionItems, err = d.GetActionItems(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load action items: %w", err)
	}

	return &r, nil
}

// GetActionItems retrieves all action items for a recording
func (d *Database) GetActionItems(recordingID int64) ([]models.ActionItem, error) {
	rows, err := d.db.Query(`
		SELECT id, recording_id, priority, task, assignee, completed, created_at
		FROM action_items WHERE recording_id = ?
		ORDER BY created_at DESC
	`, recordingID)

	if err != nil {
		return nil, fmt.Errorf("failed to query action items: %w", err)
	}
	defer rows.Close()

	var items []models.ActionItem
	for rows.Next() {
		var item models.ActionItem
		if err := rows.Scan(&item.ID, &item.RecordingID, &item.Priority,
			&item.Task, &item.Assignee, &item.Completed, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan action item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// ListRecordings retrieves all recordings, ordered by creation date
func (d *Database) ListRecordings(limit, offset int) ([]models.Recording, error) {
	rows, err := d.db.Query(`
		SELECT id, title, audio_path, duration, created_at, updated_at,
			transcript, language, summary, key_points, tags, status
		FROM recordings
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query recordings: %w", err)
	}
	defer rows.Close()

	var recordings []models.Recording
	for rows.Next() {
		var r models.Recording
		var keyPointsJSON, tagsJSON string

		if err := rows.Scan(&r.ID, &r.Title, &r.AudioPath, &r.Duration,
			&r.CreatedAt, &r.UpdatedAt, &r.Transcript, &r.Language,
			&r.Summary, &keyPointsJSON, &tagsJSON, &r.Status); err != nil {
			return nil, fmt.Errorf("failed to scan recording: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(keyPointsJSON), &r.KeyPoints); err != nil {
			r.KeyPoints = []string{}
		}
		if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
			r.Tags = []string{}
		}

		recordings = append(recordings, r)
	}

	return recordings, nil
}

// DeleteRecording deletes a recording and its action items
func (d *Database) DeleteRecording(id int64) error {
	_, err := d.db.Exec("DELETE FROM recordings WHERE id = ?", id)
	return err
}

// UpdateActionItemStatus updates the completion status of an action item
func (d *Database) UpdateActionItemStatus(id int64, completed bool) error {
	_, err := d.db.Exec("UPDATE action_items SET completed = ? WHERE id = ?", completed, id)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
