package tui

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/storo/scribescli/internal/storage"
	"github.com/storo/scribescli/pkg/models"
)

// setupTestModel creates a model with an in-memory database for testing
func setupTestModel(t *testing.T) *Model {
	t.Helper()

	db, err := storage.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return NewModel(db)
}

// setupTestModelWithRecordings creates a model and populates it with test recordings
func setupTestModelWithRecordings(t *testing.T, count int) (*Model, []models.Recording) {
	t.Helper()

	m := setupTestModel(t)
	recordings := make([]models.Recording, count)

	// Create test recordings with different timestamps
	baseTime := time.Date(2025, 11, 14, 10, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		recording := models.Recording{
			Title:      fmt.Sprintf("Test Recording %d", i+1),
			AudioPath:  filepath.Join("data", fmt.Sprintf("test_recording_%d.wav", i+1)),
			Duration:   int64((i + 1) * 60), // 60, 120, 180 seconds etc.
			CreatedAt:  baseTime.Add(time.Duration(i) * time.Hour),
			Status:     "completed",
			Transcript: fmt.Sprintf("This is test transcript %d", i+1),
			Language:   "en",
			Summary:    "",
			KeyPoints:  []string{},
			Tags:       []string{"test"},
		}

		if err := m.db.SaveRecording(&recording); err != nil {
			t.Fatalf("Failed to create test recording %d: %v", i+1, err)
		}

		recordings[i] = recording
	}

	return m, recordings
}

// TestSaveRecordingAfterStop tests that recordings are saved to database when stopped
func TestSaveRecordingAfterStop(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		audioPath      string
		duration       int64
		wantTranscript string
		wantSummary    string
		wantStatus     string
	}{
		{
			name:           "save basic recording",
			title:          "Team Standup",
			audioPath:      "data/recording_20251114_100000.wav",
			duration:       300, // 5 minutes
			wantTranscript: "",
			wantSummary:    "",
			wantStatus:     "completed",
		},
		{
			name:           "save recording with transcript",
			title:          "Client Meeting",
			audioPath:      "data/recording_20251114_110000.wav",
			duration:       1800, // 30 minutes
			wantTranscript: "This is a test transcript",
			wantSummary:    "",
			wantStatus:     "processing",
		},
		{
			name:           "save short recording",
			title:          "Quick Note",
			audioPath:      "data/recording_20251114_120000.wav",
			duration:       30, // 30 seconds
			wantTranscript: "",
			wantSummary:    "",
			wantStatus:     "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupTestModel(t)

			// Create a recording as if it was just stopped
			now := time.Now()
			recording := models.Recording{
				Title:      tt.title,
				AudioPath:  tt.audioPath,
				Duration:   tt.duration,
				CreatedAt:  now,
				UpdatedAt:  now,
				Transcript: tt.wantTranscript,
				Summary:    tt.wantSummary,
				Status:     tt.wantStatus,
				KeyPoints:  []string{},
				Tags:       []string{},
			}

			// Save the recording
			err := m.db.SaveRecording(&recording)
			if err != nil {
				t.Fatalf("Failed to save recording: %v", err)
			}

			// Verify the recording was saved with an ID
			if recording.ID == 0 {
				t.Error("Expected recording to have non-zero ID after save")
			}

			// Retrieve the recording from database
			retrieved, err := m.db.GetRecording(recording.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve recording: %v", err)
			}

			// Verify all fields match
			if retrieved.Title != tt.title {
				t.Errorf("Title mismatch: got %q, want %q", retrieved.Title, tt.title)
			}
			if retrieved.AudioPath != tt.audioPath {
				t.Errorf("AudioPath mismatch: got %q, want %q", retrieved.AudioPath, tt.audioPath)
			}
			if retrieved.Duration != tt.duration {
				t.Errorf("Duration mismatch: got %d, want %d", retrieved.Duration, tt.duration)
			}
			if retrieved.Transcript != tt.wantTranscript {
				t.Errorf("Transcript mismatch: got %q, want %q", retrieved.Transcript, tt.wantTranscript)
			}
			if retrieved.Summary != tt.wantSummary {
				t.Errorf("Summary mismatch: got %q, want %q", retrieved.Summary, tt.wantSummary)
			}
			if retrieved.Status != tt.wantStatus {
				t.Errorf("Status mismatch: got %q, want %q", retrieved.Status, tt.wantStatus)
			}

			// Verify timestamps are set
			if retrieved.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
			if retrieved.UpdatedAt.IsZero() {
				t.Error("Expected UpdatedAt to be set")
			}

			// Verify empty arrays are initialized
			if retrieved.KeyPoints == nil {
				t.Error("Expected KeyPoints to be initialized (empty array, not nil)")
			}
			if retrieved.Tags == nil {
				t.Error("Expected Tags to be initialized (empty array, not nil)")
			}
		})
	}
}

// TestLoadRecordingHistory tests loading recordings from database
func TestLoadRecordingHistory(t *testing.T) {
	tests := []struct {
		name           string
		recordingCount int
		wantMinCount   int
	}{
		{
			name:           "load empty history",
			recordingCount: 0,
			wantMinCount:   0,
		},
		{
			name:           "load single recording",
			recordingCount: 1,
			wantMinCount:   1,
		},
		{
			name:           "load multiple recordings",
			recordingCount: 5,
			wantMinCount:   5,
		},
		{
			name:           "load many recordings",
			recordingCount: 25,
			wantMinCount:   25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, createdRecordings := setupTestModelWithRecordings(t, tt.recordingCount)

			// Load recordings from database with no pagination
			recordings, err := m.db.ListRecordings(100, 0)
			if err != nil {
				t.Fatalf("Failed to load recordings: %v", err)
			}

			// Verify count
			if len(recordings) < tt.wantMinCount {
				t.Errorf("Expected at least %d recordings, got %d", tt.wantMinCount, len(recordings))
			}

			// Verify recordings are sorted by creation date (newest first)
			if len(recordings) > 1 {
				for i := 0; i < len(recordings)-1; i++ {
					if recordings[i].CreatedAt.Before(recordings[i+1].CreatedAt) {
						t.Errorf("Recordings not sorted by creation date DESC at index %d", i)
					}
				}
			}

			// Verify each recording has required fields
			for i, r := range recordings {
				if r.ID == 0 {
					t.Errorf("Recording %d has zero ID", i)
				}
				if r.Title == "" {
					t.Errorf("Recording %d has empty title", i)
				}
				if r.AudioPath == "" {
					t.Errorf("Recording %d has empty audio path", i)
				}
				if r.CreatedAt.IsZero() {
					t.Errorf("Recording %d has zero CreatedAt", i)
				}
			}

			// Verify we got the expected recordings (if any were created)
			if tt.recordingCount > 0 {
				// Check that we can find at least the first created recording
				found := false
				for _, r := range recordings {
					if r.Title == createdRecordings[0].Title {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected to find first created recording in list")
				}
			}
		})
	}
}

// TestHistoryPagination tests pagination of recording history
func TestHistoryPagination(t *testing.T) {
	tests := []struct {
		name            string
		totalRecordings int
		pageSize        int
		pageNum         int
		wantCount       int
		wantFirstTitle  string
	}{
		{
			name:            "first page of 10",
			totalRecordings: 25,
			pageSize:        10,
			pageNum:         0,
			wantCount:       10,
			wantFirstTitle:  "Test Recording 25", // Newest first
		},
		{
			name:            "second page of 10",
			totalRecordings: 25,
			pageSize:        10,
			pageNum:         1,
			wantCount:       10,
			wantFirstTitle:  "Test Recording 15",
		},
		{
			name:            "third page of 10",
			totalRecordings: 25,
			pageSize:        10,
			pageNum:         2,
			wantCount:       5, // Only 5 remaining
			wantFirstTitle:  "Test Recording 5",
		},
		{
			name:            "page beyond data",
			totalRecordings: 25,
			pageSize:        10,
			pageNum:         5,
			wantCount:       0,
			wantFirstTitle:  "",
		},
		{
			name:            "small page size",
			totalRecordings: 10,
			pageSize:        3,
			pageNum:         0,
			wantCount:       3,
			wantFirstTitle:  "Test Recording 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := setupTestModelWithRecordings(t, tt.totalRecordings)

			// Calculate offset
			offset := tt.pageNum * tt.pageSize

			// Load page
			recordings, err := m.db.ListRecordings(tt.pageSize, offset)
			if err != nil {
				t.Fatalf("Failed to load page: %v", err)
			}

			// Verify count
			if len(recordings) != tt.wantCount {
				t.Errorf("Expected %d recordings, got %d", tt.wantCount, len(recordings))
			}

			// Verify first title if we expect results
			if tt.wantCount > 0 && tt.wantFirstTitle != "" {
				if recordings[0].Title != tt.wantFirstTitle {
					t.Errorf("First recording title = %q, want %q", recordings[0].Title, tt.wantFirstTitle)
				}
			}

			// Verify pagination doesn't overlap
			if tt.pageNum > 0 && len(recordings) > 0 {
				// Get previous page
				prevRecordings, err := m.db.ListRecordings(tt.pageSize, offset-tt.pageSize)
				if err != nil {
					t.Fatalf("Failed to load previous page: %v", err)
				}

				// Ensure no overlap in IDs
				if len(prevRecordings) > 0 {
					for _, r1 := range recordings {
						for _, r2 := range prevRecordings {
							if r1.ID == r2.ID {
								t.Errorf("Found duplicate recording ID %d across pages", r1.ID)
							}
						}
					}
				}
			}
		})
	}
}

// TestSaveAnalysisWithActionItems tests saving analysis results with action items
func TestSaveAnalysisWithActionItems(t *testing.T) {
	tests := []struct {
		name        string
		summary     string
		keyPoints   []string
		actionItems []models.ActionItem
	}{
		{
			name:    "analysis with no action items",
			summary: "Team discussed project timeline and deliverables.",
			keyPoints: []string{
				"Project deadline set for Q4",
				"Need additional resources",
				"Budget approved",
			},
			actionItems: []models.ActionItem{},
		},
		{
			name:    "analysis with single action item",
			summary: "Sprint planning meeting for upcoming iteration.",
			keyPoints: []string{
				"Sprint goal defined",
				"User stories prioritized",
			},
			actionItems: []models.ActionItem{
				{
					Priority:  "HIGH",
					Task:      "Update sprint backlog in Jira",
					Assignee:  "John Doe",
					Completed: false,
				},
			},
		},
		{
			name:    "analysis with multiple action items",
			summary: "Quarterly business review with stakeholders.",
			keyPoints: []string{
				"Revenue exceeded targets by 15%",
				"Customer satisfaction improved",
				"Need to address support ticket backlog",
			},
			actionItems: []models.ActionItem{
				{
					Priority:  "HIGH",
					Task:      "Hire 2 additional support engineers",
					Assignee:  "HR Team",
					Completed: false,
				},
				{
					Priority:  "MEDIUM",
					Task:      "Prepare Q4 budget proposal",
					Assignee:  "Finance",
					Completed: false,
				},
				{
					Priority:  "LOW",
					Task:      "Schedule team celebration",
					Assignee:  "Admin",
					Completed: false,
				},
			},
		},
		{
			name:    "analysis with completed action items",
			summary: "Follow-up meeting on action items.",
			keyPoints: []string{
				"Most items completed",
				"Some items require more time",
			},
			actionItems: []models.ActionItem{
				{
					Priority:  "HIGH",
					Task:      "Deploy hotfix to production",
					Assignee:  "DevOps",
					Completed: true,
				},
				{
					Priority:  "MEDIUM",
					Task:      "Update documentation",
					Assignee:  "Tech Writer",
					Completed: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupTestModel(t)

			// Create a recording with analysis results
			recording := models.Recording{
				Title:       "Test Meeting",
				AudioPath:   "data/test.wav",
				Duration:    1800,
				CreatedAt:   time.Now(),
				Status:      "completed",
				Transcript:  "Meeting transcript here",
				Summary:     tt.summary,
				KeyPoints:   tt.keyPoints,
				ActionItems: tt.actionItems,
				Tags:        []string{"test"},
			}

			// Save the recording (should cascade save action items)
			err := m.db.SaveRecording(&recording)
			if err != nil {
				t.Fatalf("Failed to save recording with analysis: %v", err)
			}

			// Retrieve the recording
			retrieved, err := m.db.GetRecording(recording.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve recording: %v", err)
			}

			// Verify summary
			if retrieved.Summary != tt.summary {
				t.Errorf("Summary mismatch: got %q, want %q", retrieved.Summary, tt.summary)
			}

			// Verify key points
			if len(retrieved.KeyPoints) != len(tt.keyPoints) {
				t.Errorf("KeyPoints count mismatch: got %d, want %d", len(retrieved.KeyPoints), len(tt.keyPoints))
			}
			for i, kp := range tt.keyPoints {
				if i < len(retrieved.KeyPoints) && retrieved.KeyPoints[i] != kp {
					t.Errorf("KeyPoint %d mismatch: got %q, want %q", i, retrieved.KeyPoints[i], kp)
				}
			}

			// Verify action items
			if len(retrieved.ActionItems) != len(tt.actionItems) {
				t.Errorf("ActionItems count mismatch: got %d, want %d", len(retrieved.ActionItems), len(tt.actionItems))
			}

			for i, expected := range tt.actionItems {
				if i >= len(retrieved.ActionItems) {
					break
				}
				actual := retrieved.ActionItems[i]

				if actual.Priority != expected.Priority {
					t.Errorf("ActionItem %d Priority mismatch: got %q, want %q", i, actual.Priority, expected.Priority)
				}
				if actual.Task != expected.Task {
					t.Errorf("ActionItem %d Task mismatch: got %q, want %q", i, actual.Task, expected.Task)
				}
				if actual.Assignee != expected.Assignee {
					t.Errorf("ActionItem %d Assignee mismatch: got %q, want %q", i, actual.Assignee, expected.Assignee)
				}
				if actual.Completed != expected.Completed {
					t.Errorf("ActionItem %d Completed mismatch: got %v, want %v", i, actual.Completed, expected.Completed)
				}
				if actual.RecordingID != recording.ID {
					t.Errorf("ActionItem %d RecordingID mismatch: got %d, want %d", i, actual.RecordingID, recording.ID)
				}
			}
		})
	}
}

// TestSelectRecordingFromHistory tests selecting and loading a recording's full details
func TestSelectRecordingFromHistory(t *testing.T) {
	m, createdRecordings := setupTestModelWithRecordings(t, 5)

	tests := []struct {
		name        string
		recordingID int64
		wantError   bool
	}{
		{
			name:        "select first recording",
			recordingID: createdRecordings[0].ID,
			wantError:   false,
		},
		{
			name:        "select middle recording",
			recordingID: createdRecordings[2].ID,
			wantError:   false,
		},
		{
			name:        "select last recording",
			recordingID: createdRecordings[4].ID,
			wantError:   false,
		},
		{
			name:        "select non-existent recording",
			recordingID: 99999,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load the full recording details
			recording, err := m.db.GetRecording(tt.recordingID)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error when loading non-existent recording, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to load recording: %v", err)
			}

			// Verify all fields are loaded
			if recording.ID != tt.recordingID {
				t.Errorf("ID mismatch: got %d, want %d", recording.ID, tt.recordingID)
			}
			if recording.Title == "" {
				t.Error("Expected title to be loaded")
			}
			if recording.AudioPath == "" {
				t.Error("Expected audio path to be loaded")
			}
			if recording.Duration == 0 {
				t.Error("Expected duration to be loaded")
			}
			if recording.CreatedAt.IsZero() {
				t.Error("Expected created_at to be loaded")
			}

			// Verify arrays are initialized (even if empty)
			if recording.KeyPoints == nil {
				t.Error("Expected KeyPoints to be initialized")
			}
			if recording.Tags == nil {
				t.Error("Expected Tags to be initialized")
			}
			if recording.ActionItems == nil {
				t.Error("Expected ActionItems to be initialized")
			}
		})
	}
}

// TestDatabaseErrorHandling tests graceful error handling for database operations
func TestDatabaseErrorHandling(t *testing.T) {
	t.Run("handle database connection error", func(t *testing.T) {
		// Try to connect to invalid database path
		_, err := storage.NewDatabase("/invalid/path/that/does/not/exist/db.sqlite")
		if err == nil {
			t.Error("Expected error when creating database with invalid path")
		}
	})

	t.Run("handle get non-existent recording", func(t *testing.T) {
		m := setupTestModel(t)

		_, err := m.db.GetRecording(99999)
		if err == nil {
			t.Error("Expected error when getting non-existent recording")
		}
	})

	t.Run("handle invalid recording data", func(t *testing.T) {
		m := setupTestModel(t)

		// Try to save recording with missing required fields
		recording := models.Recording{
			// Missing Title, AudioPath
			Duration:  100,
			CreatedAt: time.Now(),
		}

		err := m.db.SaveRecording(&recording)
		// Note: SQLite might allow this, but we test the behavior
		// In a real scenario, you might add validation before saving
		if err != nil {
			// Error is acceptable - validation caught it
			t.Logf("Validation error (expected): %v", err)
		} else {
			// SQLite allowed it - verify we can still retrieve it
			retrieved, err := m.db.GetRecording(recording.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve recording with partial data: %v", err)
			}
			if retrieved.Title != "" {
				t.Error("Expected empty title for recording with missing data")
			}
		}
	})

	t.Run("handle delete non-existent recording", func(t *testing.T) {
		m := setupTestModel(t)

		// Deleting non-existent recording should not error (idempotent)
		err := m.db.DeleteRecording(99999)
		if err != nil {
			t.Errorf("Expected no error when deleting non-existent recording, got: %v", err)
		}
	})

	t.Run("handle cascade delete action items", func(t *testing.T) {
		m := setupTestModel(t)

		// Create recording with action items
		recording := models.Recording{
			Title:     "Test Meeting",
			AudioPath: "data/test.wav",
			Duration:  300,
			Status:    "completed",
			ActionItems: []models.ActionItem{
				{
					Priority: "HIGH",
					Task:     "Test task",
					Assignee: "Test user",
				},
			},
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		// Verify action item was created
		items, err := m.db.GetActionItems(recording.ID)
		if err != nil {
			t.Fatalf("Failed to get action items: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("Expected 1 action item, got %d", len(items))
		}

		// Delete the recording
		err = m.db.DeleteRecording(recording.ID)
		if err != nil {
			t.Fatalf("Failed to delete recording: %v", err)
		}

		// Verify action items were cascade deleted
		items, err = m.db.GetActionItems(recording.ID)
		if err != nil {
			t.Fatalf("Failed to get action items after delete: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 action items after cascade delete, got %d", len(items))
		}
	})
}

// TestRecordingTimestamps tests that timestamps are properly managed
func TestRecordingTimestamps(t *testing.T) {
	t.Run("auto-set created_at when zero", func(t *testing.T) {
		m := setupTestModel(t)

		beforeSave := time.Now()
		recording := models.Recording{
			Title:     "Test",
			AudioPath: "test.wav",
			Duration:  100,
			Status:    "completed",
			// CreatedAt not set (zero value)
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		afterSave := time.Now()

		// Verify CreatedAt was auto-set
		if recording.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be auto-set when zero")
		}

		// Verify CreatedAt is within reasonable range
		if recording.CreatedAt.Before(beforeSave) || recording.CreatedAt.After(afterSave) {
			t.Errorf("CreatedAt %v not between %v and %v", recording.CreatedAt, beforeSave, afterSave)
		}
	})

	t.Run("preserve explicit created_at", func(t *testing.T) {
		m := setupTestModel(t)

		explicitTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		recording := models.Recording{
			Title:     "Test",
			AudioPath: "test.wav",
			Duration:  100,
			Status:    "completed",
			CreatedAt: explicitTime,
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		// Verify explicit CreatedAt was preserved
		if !recording.CreatedAt.Equal(explicitTime) {
			t.Errorf("CreatedAt = %v, want %v", recording.CreatedAt, explicitTime)
		}
	})

	t.Run("always set updated_at", func(t *testing.T) {
		m := setupTestModel(t)

		beforeSave := time.Now()
		recording := models.Recording{
			Title:     "Test",
			AudioPath: "test.wav",
			Duration:  100,
			Status:    "completed",
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		afterSave := time.Now()

		// Verify UpdatedAt was set
		if recording.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}

		// Verify UpdatedAt is within reasonable range
		if recording.UpdatedAt.Before(beforeSave) || recording.UpdatedAt.After(afterSave) {
			t.Errorf("UpdatedAt %v not between %v and %v", recording.UpdatedAt, beforeSave, afterSave)
		}
	})
}

// TestActionItemOperations tests individual action item operations
func TestActionItemOperations(t *testing.T) {
	t.Run("update action item completion status", func(t *testing.T) {
		m := setupTestModel(t)

		// Create a recording with an action item
		recording := models.Recording{
			Title:     "Test Meeting",
			AudioPath: "test.wav",
			Duration:  300,
			Status:    "completed",
			ActionItems: []models.ActionItem{
				{
					Priority:  "HIGH",
					Task:      "Complete report",
					Assignee:  "John",
					Completed: false,
				},
			},
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		itemID := recording.ActionItems[0].ID

		// Update completion status
		err = m.db.UpdateActionItemStatus(itemID, true)
		if err != nil {
			t.Fatalf("Failed to update action item status: %v", err)
		}

		// Verify the update
		items, err := m.db.GetActionItems(recording.ID)
		if err != nil {
			t.Fatalf("Failed to get action items: %v", err)
		}

		if len(items) != 1 {
			t.Fatalf("Expected 1 action item, got %d", len(items))
		}

		if !items[0].Completed {
			t.Error("Expected action item to be marked as completed")
		}

		// Update back to incomplete
		err = m.db.UpdateActionItemStatus(itemID, false)
		if err != nil {
			t.Fatalf("Failed to update action item status: %v", err)
		}

		// Verify the update
		items, err = m.db.GetActionItems(recording.ID)
		if err != nil {
			t.Fatalf("Failed to get action items: %v", err)
		}

		if items[0].Completed {
			t.Error("Expected action item to be marked as incomplete")
		}
	})

	t.Run("get action items ordered by creation", func(t *testing.T) {
		m := setupTestModel(t)

		// Create a recording with multiple action items
		recording := models.Recording{
			Title:     "Test Meeting",
			AudioPath: "test.wav",
			Duration:  300,
			Status:    "completed",
			ActionItems: []models.ActionItem{
				{Priority: "HIGH", Task: "First task", Assignee: "Alice"},
				{Priority: "MEDIUM", Task: "Second task", Assignee: "Bob"},
				{Priority: "LOW", Task: "Third task", Assignee: "Charlie"},
			},
		}

		err := m.db.SaveRecording(&recording)
		if err != nil {
			t.Fatalf("Failed to save recording: %v", err)
		}

		// Get action items
		items, err := m.db.GetActionItems(recording.ID)
		if err != nil {
			t.Fatalf("Failed to get action items: %v", err)
		}

		if len(items) != 3 {
			t.Fatalf("Expected 3 action items, got %d", len(items))
		}

		// Verify they're ordered by creation date (DESC)
		for i := 0; i < len(items)-1; i++ {
			if items[i].CreatedAt.Before(items[i+1].CreatedAt) {
				t.Errorf("Action items not ordered by creation date DESC at index %d", i)
			}
		}
	})
}
