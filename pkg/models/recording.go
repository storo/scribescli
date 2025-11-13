package models

import (
	"time"
)

// Recording represents a single meeting recording
type Recording struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	AudioPath   string    `json:"audio_path"`
	Duration    int64     `json:"duration"` // in seconds
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Transcription
	Transcript  string    `json:"transcript"`
	Language    string    `json:"language"`

	// AI Analysis
	Summary     string    `json:"summary"`
	KeyPoints   []string  `json:"key_points"`
	ActionItems []ActionItem `json:"action_items"`

	// Metadata
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"` // "recording", "processing", "completed", "failed"
}

// ActionItem represents an actionable task from a meeting
type ActionItem struct {
	ID         int64     `json:"id"`
	RecordingID int64    `json:"recording_id"`
	Priority   string    `json:"priority"` // "HIGH", "MEDIUM", "LOW"
	Task       string    `json:"task"`
	Assignee   string    `json:"assignee"`
	Completed  bool      `json:"completed"`
	CreatedAt  time.Time `json:"created_at"`
}

// TranscriptSegment represents a timestamped segment of transcript
type TranscriptSegment struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Speaker   string  `json:"speaker"` // Optional speaker identification
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatText     ExportFormat = "text"
	ExportFormatPDF      ExportFormat = "pdf"
)
