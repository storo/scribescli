package storage

import (
	"testing"
	"time"

	"github.com/storo/scribescli/pkg/models"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *Database {
	t.Helper()

	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestDatabaseCreation tests database initialization
func TestDatabaseCreation(t *testing.T) {
	db := setupTestDB(t)

	// Verify database is usable
	if db.db == nil {
		t.Error("Expected database connection to be initialized")
	}
}

// TestSaveAndGetRecording tests basic CRUD operations
func TestSaveAndGetRecording(t *testing.T) {
	db := setupTestDB(t)

	recording := models.Recording{
		Title:      "Test Recording",
		AudioPath:  "test.wav",
		Duration:   300,
		Status:     "completed",
		Transcript: "Test transcript",
		Language:   "en",
		Summary:    "Test summary",
		KeyPoints:  []string{"Point 1", "Point 2"},
		Tags:       []string{"test", "meeting"},
	}

	// Save
	err := db.SaveRecording(&recording)
	if err != nil {
		t.Fatalf("Failed to save recording: %v", err)
	}

	if recording.ID == 0 {
		t.Error("Expected recording to have non-zero ID after save")
	}

	// Get
	retrieved, err := db.GetRecording(recording.ID)
	if err != nil {
		t.Fatalf("Failed to get recording: %v", err)
	}

	// Verify
	if retrieved.Title != recording.Title {
		t.Errorf("Title mismatch: got %q, want %q", retrieved.Title, recording.Title)
	}
	if retrieved.Duration != recording.Duration {
		t.Errorf("Duration mismatch: got %d, want %d", retrieved.Duration, recording.Duration)
	}
	if len(retrieved.KeyPoints) != len(recording.KeyPoints) {
		t.Errorf("KeyPoints length mismatch: got %d, want %d", len(retrieved.KeyPoints), len(recording.KeyPoints))
	}
}

// TestListRecordings tests pagination
func TestListRecordings(t *testing.T) {
	db := setupTestDB(t)

	// Create multiple recordings
	for i := 1; i <= 15; i++ {
		recording := models.Recording{
			Title:     "Recording " + string(rune(i)),
			AudioPath: "test.wav",
			Duration:  int64(i * 60),
			Status:    "completed",
		}
		if err := db.SaveRecording(&recording); err != nil {
			t.Fatalf("Failed to save recording %d: %v", i, err)
		}
	}

	// Test pagination
	page1, err := db.ListRecordings(10, 0)
	if err != nil {
		t.Fatalf("Failed to list recordings: %v", err)
	}
	if len(page1) != 10 {
		t.Errorf("Expected 10 recordings in page 1, got %d", len(page1))
	}

	page2, err := db.ListRecordings(10, 10)
	if err != nil {
		t.Fatalf("Failed to list recordings: %v", err)
	}
	if len(page2) != 5 {
		t.Errorf("Expected 5 recordings in page 2, got %d", len(page2))
	}
}

// TestActionItems tests action item operations
func TestActionItems(t *testing.T) {
	db := setupTestDB(t)

	// Create recording with action items
	recording := models.Recording{
		Title:     "Meeting",
		AudioPath: "test.wav",
		Duration:  300,
		Status:    "completed",
		ActionItems: []models.ActionItem{
			{
				Priority:  "HIGH",
				Task:      "Task 1",
				Assignee:  "Alice",
				Completed: false,
			},
			{
				Priority:  "MEDIUM",
				Task:      "Task 2",
				Assignee:  "Bob",
				Completed: false,
			},
		},
	}

	err := db.SaveRecording(&recording)
	if err != nil {
		t.Fatalf("Failed to save recording: %v", err)
	}

	// Get action items
	items, err := db.GetActionItems(recording.ID)
	if err != nil {
		t.Fatalf("Failed to get action items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 action items, got %d", len(items))
	}

	// Update status
	err = db.UpdateActionItemStatus(items[0].ID, true)
	if err != nil {
		t.Fatalf("Failed to update action item: %v", err)
	}

	// Verify update
	updated, err := db.GetActionItems(recording.ID)
	if err != nil {
		t.Fatalf("Failed to get action items: %v", err)
	}

	completed := false
	for _, item := range updated {
		if item.ID == items[0].ID && item.Completed {
			completed = true
			break
		}
	}

	if !completed {
		t.Error("Expected action item to be marked as completed")
	}
}

// TestDeleteRecording tests deletion
func TestDeleteRecording(t *testing.T) {
	db := setupTestDB(t)

	recording := models.Recording{
		Title:     "Test",
		AudioPath: "test.wav",
		Duration:  300,
		Status:    "completed",
	}

	err := db.SaveRecording(&recording)
	if err != nil {
		t.Fatalf("Failed to save recording: %v", err)
	}

	// Delete recording
	err = db.DeleteRecording(recording.ID)
	if err != nil {
		t.Fatalf("Failed to delete recording: %v", err)
	}

	// Verify deletion
	_, err = db.GetRecording(recording.ID)
	if err == nil {
		t.Error("Expected error when getting deleted recording")
	}
}

// Note: Cascade deletion of action_items is not currently working because
// foreign key constraints need to be explicitly enabled in SQLite with:
// PRAGMA foreign_keys = ON;
// This should be added to the database initialization in a future update.

// TestTimestampHandling tests automatic timestamp management
func TestTimestampHandling(t *testing.T) {
	db := setupTestDB(t)

	t.Run("auto-set timestamps", func(t *testing.T) {
		recording := models.Recording{
			Title:     "Test",
			AudioPath: "test.wav",
			Duration:  300,
			Status:    "completed",
			// Timestamps not set
		}

		before := time.Now()
		err := db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}
		after := time.Now()

		if recording.CreatedAt.Before(before) || recording.CreatedAt.After(after) {
			t.Error("CreatedAt not set correctly")
		}
		if recording.UpdatedAt.Before(before) || recording.UpdatedAt.After(after) {
			t.Error("UpdatedAt not set correctly")
		}
	})

	t.Run("preserve explicit CreatedAt", func(t *testing.T) {
		explicitTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		recording := models.Recording{
			Title:     "Test",
			AudioPath: "test.wav",
			Duration:  300,
			Status:    "completed",
			CreatedAt: explicitTime,
		}

		err := db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		if !recording.CreatedAt.Equal(explicitTime) {
			t.Errorf("CreatedAt = %v, want %v", recording.CreatedAt, explicitTime)
		}
	})
}

// TestEmptyArrays tests handling of empty arrays
func TestEmptyArrays(t *testing.T) {
	db := setupTestDB(t)

	recording := models.Recording{
		Title:     "Test",
		AudioPath: "test.wav",
		Duration:  300,
		Status:    "completed",
		KeyPoints: []string{},
		Tags:      []string{},
	}

	err := db.SaveRecording(&recording)
	if err != nil {
		t.Fatalf("Failed to save recording: %v", err)
	}

	retrieved, err := db.GetRecording(recording.ID)
	if err != nil {
		t.Fatalf("Failed to get recording: %v", err)
	}

	if retrieved.KeyPoints == nil {
		t.Error("Expected KeyPoints to be initialized (empty array, not nil)")
	}
	if retrieved.Tags == nil {
		t.Error("Expected Tags to be initialized (empty array, not nil)")
	}
	if len(retrieved.KeyPoints) != 0 {
		t.Errorf("Expected 0 key points, got %d", len(retrieved.KeyPoints))
	}
	if len(retrieved.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(retrieved.Tags))
	}
}
