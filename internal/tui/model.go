package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/storo/scribescli/internal/ai"
	"github.com/storo/scribescli/internal/audio"
	"github.com/storo/scribescli/internal/export"
	"github.com/storo/scribescli/internal/storage"
	"github.com/storo/scribescli/internal/transcription"
	"github.com/storo/scribescli/pkg/models"
)

// ViewState represents different views in the application
type ViewState int

const (
	ViewMenu ViewState = iota
	ViewRecording
	ViewProcessing
	ViewAnalysis
	ViewHistory
	ViewSettings
)

// Model represents the main application model
type Model struct {
	// Current view state
	view ViewState

	// Components
	halEye      *HALEye
	recorder    *audio.Recorder
	db          *storage.Database
	transcriber *transcription.VoskTranscriber
	modelMgr    *transcription.ModelManager
	aiClient    *ai.ClaudeClient

	// Recording state
	recording         bool
	paused            bool
	recordingTime     time.Duration
	audioLevel        float32
	transcript        []string
	currentTranscript string

	// Analysis results
	summary     string
	keyPoints   []string
	actionItems []ActionItem

	// Menu state
	menuCursor int
	menuItems  []string

	// Database integration
	currentRecordingID int64
	recordings         []models.Recording
	historyOffset      int
	historyLimit       int

	// Window size
	width  int
	height int

	// Error message
	err error

	// Processing state
	processingMessage string
	processingError   error

	// Retry state
	retryCount int
	maxRetries int

	// Keybindings
	keys KeyMap
}

// ActionItem represents an actionable task from the meeting
type ActionItem struct {
	Priority string
	Task     string
	Assignee string
}

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Record key.Binding
	Stop   key.Binding
	Pause  key.Binding
	Back   key.Binding
	Quit   key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Record: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "record"),
		),
		Stop: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "stop"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause/resume"),
		),
		Back: key.NewBinding(
			key.WithKeys("b", "esc"),
			key.WithHelp("b/esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// NewModel creates a new application model
func NewModel(db *storage.Database) *Model {
	// Don't initialize recorder yet - do it lazily when needed
	// Initialize model manager for Vosk models
	modelMgr := transcription.NewModelManager("./models")

	// Initialize AI client
	aiClient, err := ai.NewClaudeClient()
	var aiClientRef *ai.ClaudeClient
	if err != nil {
		// Don't fail app startup - just log warning
		// Will show error when user tries to analyze
		aiClientRef = nil
	} else {
		aiClientRef = aiClient
	}

	return &Model{
		view:        ViewMenu,
		halEye:      NewHALEye(),
		recorder:    nil, // Initialize lazily
		db:          db,
		transcriber: nil, // Initialize lazily with model
		modelMgr:    modelMgr,
		aiClient:    aiClientRef,
		err:         nil,
		menuItems: []string{
			"New Recording",
			"History",
			"Settings",
			"Quit",
		},
		keys:          DefaultKeyMap(),
		historyOffset: 0,
		historyLimit:  10,
		recordings:    nil, // Load lazily
		maxRetries:    3,
		retryCount:    0,
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tickHALEye(),
		tea.EnterAltScreen,
	)
}

// TickMsg is sent on each animation tick
type TickMsg time.Time

// tickHALEye returns a command that ticks the HAL eye animation
func tickHALEye() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// RecordingTickMsg is sent to update recording time and audio level
type RecordingTickMsg time.Time

// tickRecording returns a command that ticks the recording updates
func tickRecording() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return RecordingTickMsg(t)
	})
}

// analyzeRecording returns a command that analyzes a recording with Claude
func analyzeRecording(aiClient *ai.ClaudeClient, recordingID int64, transcript string) tea.Cmd {
	return func() tea.Msg {
		// Check for empty transcript
		if strings.TrimSpace(transcript) == "" {
			return AnalysisCompleteMsg{
				RecordingID: recordingID,
				Summary:     "No transcript available for analysis.",
				KeyPoints:   []string{},
				ActionItems: []ActionItem{},
				Err:         nil,
			}
		}

		// Call Claude API with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		analysis, err := aiClient.AnalyzeMeeting(ctx, transcript)
		if err != nil {
			return AnalysisErrorMsg{
				RecordingID: recordingID,
				Err:         fmt.Errorf("Claude API error: %w", err),
			}
		}

		// Convert ai.ActionItem to tui.ActionItem
		actionItems := make([]ActionItem, len(analysis.ActionItems))
		for i, item := range analysis.ActionItems {
			actionItems[i] = ActionItem{
				Priority: item.Priority,
				Task:     item.Task,
				Assignee: item.Assignee,
			}
		}

		return AnalysisCompleteMsg{
			RecordingID: recordingID,
			Summary:     analysis.Summary,
			KeyPoints:   analysis.KeyPoints,
			ActionItems: actionItems,
			Err:         nil,
		}
	}
}

// TranscriptMsg is sent when transcription results are available
type TranscriptMsg struct {
	Text    string
	IsFinal bool
}

// AnalysisCompleteMsg is sent when Claude analysis finishes successfully
type AnalysisCompleteMsg struct {
	RecordingID int64
	Summary     string
	KeyPoints   []string
	ActionItems []ActionItem
	Err         error
}

// AnalysisErrorMsg is sent when analysis fails
type AnalysisErrorMsg struct {
	RecordingID int64
	Err         error
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		m.halEye.Update()
		return m, tickHALEye()

	case RecordingTickMsg:
		if m.recording && m.recorder != nil {
			m.recordingTime = m.recorder.GetDuration()
			m.audioLevel = m.recorder.GetAudioLevel()
			return m, tickRecording()
		}
		return m, nil

	case TranscriptMsg:
		// Update transcript display
		if msg.IsFinal {
			// Final result - add to transcript history
			m.transcript = append(m.transcript, msg.Text)
			m.currentTranscript = "" // Clear partial
		} else {
			// Partial result - update current line
			m.currentTranscript = msg.Text
		}
		return m, nil

	case AnalysisCompleteMsg:
		// Reset retry count on success
		m.retryCount = 0

		// Analysis succeeded
		if msg.Err != nil {
			m.processingError = msg.Err
			m.err = fmt.Errorf("Analysis completed with errors: %w", msg.Err)
			m.view = ViewMenu
			m.halEye.SetState("idle")
			return m, nil
		}

		// Save analysis results to database
		if err := m.saveAnalysis(msg.RecordingID, msg.Summary, msg.KeyPoints, msg.ActionItems); err != nil {
			m.err = fmt.Errorf("Failed to save analysis: %w", err)
			m.view = ViewMenu
			m.halEye.SetState("idle")
			return m, nil
		}

		// Update model state for display
		m.summary = msg.Summary
		m.keyPoints = msg.KeyPoints
		m.actionItems = msg.ActionItems

		// Transition to analysis view
		m.view = ViewAnalysis
		m.halEye.SetState("idle")
		m.err = nil
		return m, nil

	case AnalysisErrorMsg:
		// Check if we should retry
		if m.retryCount < m.maxRetries {
			m.retryCount++
			m.processingMessage = fmt.Sprintf("Retrying analysis... (attempt %d/%d)",
				m.retryCount, m.maxRetries)

			// Reload recording to get transcript
			recording, err := m.db.GetRecording(msg.RecordingID)
			if err != nil {
				m.err = fmt.Errorf("Failed to retry analysis: %w", err)
				m.view = ViewMenu
				m.halEye.SetState("error")
				m.retryCount = 0
				return m, nil
			}

			// Retry analysis
			return m, analyzeRecording(m.aiClient, msg.RecordingID, recording.Transcript)
		}

		// Max retries exceeded
		m.retryCount = 0
		m.processingError = msg.Err
		m.err = fmt.Errorf("Analysis failed after %d attempts: %w. Recording saved without analysis.",
			m.maxRetries, msg.Err)
		m.view = ViewMenu
		m.halEye.SetState("error")
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.recorder != nil {
				m.recorder.Close()
			}
			if m.transcriber != nil {
				m.transcriber.Close()
			}
			return m, tea.Quit

		case key.Matches(msg, m.keys.Back):
			if m.view != ViewMenu {
				m.view = ViewMenu
				m.halEye.SetState("idle")
				return m, nil
			}
		}

		// View-specific key handling
		switch m.view {
		case ViewMenu:
			return m.updateMenu(msg)
		case ViewRecording:
			return m.updateRecording(msg)
		case ViewAnalysis:
			return m.updateAnalysis(msg)
		case ViewHistory:
			return m.updateHistory(msg)
		}
	}

	return m, nil
}

// updateMenu handles menu view updates
func (m *Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.menuCursor > 0 {
			m.menuCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.menuCursor < len(m.menuItems)-1 {
			m.menuCursor++
		}

	case key.Matches(msg, m.keys.Enter):
		switch m.menuCursor {
		case 0: // New Recording
			// Initialize recorder lazily (only when user wants to record)
			if m.recorder == nil {
				recorder, err := audio.NewRecorder()
				if err != nil {
					m.err = fmt.Errorf("Cannot initialize audio: %v\n\nPlease ensure PortAudio is installed correctly", err)
					return m, nil
				}
				m.recorder = recorder
			}

			// Initialize transcriber if not already done
			if m.transcriber == nil {
				// Try to get default model path
				modelPath, err := m.modelMgr.GetModelPath(m.modelMgr.GetDefaultModel())
				if err != nil {
					// Model not found - try to download default model
					// For now, proceed without transcription (graceful degradation)
					// TODO: Add proper model download UI with progress
					m.err = fmt.Errorf("Vosk model not found. Recording without transcription.\n\nRun: ./scripts/install-vosk.sh && download model")
				} else {
					// Initialize transcriber with model
					transcriber := transcription.NewVoskTranscriber()
					if err := transcriber.Initialize(modelPath); err != nil {
						m.err = fmt.Errorf("Cannot initialize transcription: %v", err)
					} else {
						m.transcriber = transcriber
					}
				}
			}

			// Clear previous transcript
			m.transcript = []string{}
			m.currentTranscript = ""

			m.view = ViewRecording
			m.halEye.SetState("recording")
			if m.recorder != nil {
				err := m.recorder.Start()
				if err != nil {
					m.err = fmt.Errorf("Cannot start recording: %v", err)
					m.view = ViewMenu
					return m, nil
				}
				m.recording = true
				return m, tickRecording()
			}
		case 1: // History
			m.view = ViewHistory
		case 2: // Settings
			m.view = ViewSettings
		case 3: // Quit
			if m.recorder != nil {
				m.recorder.Close()
			}
			if m.transcriber != nil {
				m.transcriber.Close()
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

// updateRecording handles recording view updates
func (m *Model) updateRecording(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Stop):
		if m.recording && m.recorder != nil {
			// 1. Stop recorder
			if err := m.recorder.Stop(); err != nil {
				m.err = err
				return m, nil
			}

			m.recording = false
			m.halEye.SetState("processing")

			// 2. Generate filename with timestamp
			timestamp := time.Now()
			filename := fmt.Sprintf("data/recording_%s.wav", timestamp.Format("20060102_150405"))

			// 3. Save WAV file
			if err := m.recorder.SaveRecording(filename); err != nil {
				m.err = fmt.Errorf("Error saving recording: %w", err)
				return m, nil
			}

			// 4. Get duration
			duration := m.recordingTime

			// 5. Build transcript from collected results
			var transcript string
			if len(m.transcript) > 0 {
				// Join all final transcript segments
				transcript = ""
				for i, segment := range m.transcript {
					if i > 0 {
						transcript += " "
					}
					transcript += segment
				}
			}

			// 6. Create recording model
			recording := &models.Recording{
				Title:      fmt.Sprintf("Meeting %s", timestamp.Format("2006-01-02 15:04")),
				AudioPath:  filename,
				Duration:   int64(duration.Seconds()),
				CreatedAt:  timestamp,
				Transcript: transcript,
				Summary:    "",
				KeyPoints:  []string{},
				Tags:       []string{"unprocessed"},
				Status:     "processing", // Will be updated to "completed" after analysis
			}

			// 7. Save to database
			if err := m.db.SaveRecording(recording); err != nil {
				m.err = fmt.Errorf("Error saving to database: %w", err)
				return m, nil
			}

			// 8. Update state
			m.currentRecordingID = recording.ID

			// 9. Check if we can analyze
			if m.aiClient == nil {
				// No AI client - skip analysis, show warning
				m.err = fmt.Errorf("Claude API not available. Recording saved without analysis.")
				m.view = ViewMenu
				m.halEye.SetState("idle")
				return m, nil
			}

			// Check if transcript is empty
			if transcript == "" {
				// No transcript - skip analysis with warning
				m.err = fmt.Errorf("No transcript available. Recording saved without analysis.")
				m.view = ViewMenu
				m.halEye.SetState("idle")
				return m, nil
			}

			// 10. Transition to processing view
			m.view = ViewProcessing
			m.halEye.SetState("processing")
			m.processingMessage = "Analyzing transcript with Claude AI..."
			m.processingError = nil

			// 11. Trigger async analysis
			return m, analyzeRecording(m.aiClient, recording.ID, transcript)
		}

	case key.Matches(msg, m.keys.Pause):
		if m.recorder != nil {
			if m.paused {
				m.recorder.Resume()
				m.paused = false
			} else {
				m.recorder.Pause()
				m.paused = true
			}
		}
	}

	return m, nil
}

// updateAnalysis handles analysis view updates
func (m *Model) updateAnalysis(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("m"))):
		// Export to Markdown
		return m.handleExport("markdown")

	case key.Matches(msg, key.NewBinding(key.WithKeys("j"))):
		// Export to JSON
		return m.handleExport("json")

	case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
		// Export to Text
		return m.handleExport("text")

	case key.Matches(msg, m.keys.Back):
		m.view = ViewMenu
		m.menuCursor = 0
		m.halEye.SetState("idle")
		return m, nil
	}

	return m, nil
}

// updateHistory handles history view updates
func (m *Model) updateHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.menuCursor > 0 {
			m.menuCursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.recordings != nil && m.menuCursor < len(m.recordings)-1 {
			m.menuCursor++
		}

	case key.Matches(msg, m.keys.Enter):
		// View recording details
		if m.recordings != nil && m.menuCursor < len(m.recordings) {
			selectedRecording := m.recordings[m.menuCursor]
			m.currentRecordingID = selectedRecording.ID

			// Load full recording with action items
			recording, err := m.db.GetRecording(selectedRecording.ID)
			if err != nil {
				m.err = fmt.Errorf("Error loading recording: %w", err)
				return m, nil
			}

			// Update model state with recording data
			m.summary = recording.Summary
			m.keyPoints = recording.KeyPoints
			m.recordingTime = time.Duration(recording.Duration) * time.Second

			// Convert models.ActionItem to tui.ActionItem
			m.actionItems = make([]ActionItem, len(recording.ActionItems))
			for i, item := range recording.ActionItems {
				m.actionItems[i] = ActionItem{
					Priority: item.Priority,
					Task:     item.Task,
					Assignee: item.Assignee,
				}
			}

			m.view = ViewAnalysis
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
		// Next page
		m.historyOffset += m.historyLimit
		m.recordings = nil // Force reload
		m.menuCursor = 0   // Reset cursor

	case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
		// Previous page
		if m.historyOffset >= m.historyLimit {
			m.historyOffset -= m.historyLimit
			m.recordings = nil // Force reload
			m.menuCursor = 0   // Reset cursor
		}

	case key.Matches(msg, m.keys.Back):
		m.view = ViewMenu
		m.menuCursor = 0
		m.recordings = nil // Clear cache
	}

	return m, nil
}

// saveAnalysis persists analysis results to database
// Will be used in Phase 4 when Claude integration is complete
func (m *Model) saveAnalysis(recordingID int64, summary string, keyPoints []string, actionItems []ActionItem) error {
	// Load recording
	recording, err := m.db.GetRecording(recordingID)
	if err != nil {
		return fmt.Errorf("error loading recording: %w", err)
	}

	// Update with analysis
	recording.Summary = summary
	recording.KeyPoints = keyPoints
	recording.Status = "completed" // Mark as completed after analysis

	// Remove "unprocessed" tag, add "analyzed"
	recording.Tags = []string{"analyzed"}

	if err := m.db.UpdateRecording(recording); err != nil {
		return fmt.Errorf("error updating recording: %w", err)
	}

	// Save action items
	for _, item := range actionItems {
		actionItem := &models.ActionItem{
			RecordingID: recordingID,
			Priority:    item.Priority,
			Task:        item.Task,
			Assignee:    item.Assignee,
		}

		if err := m.db.SaveActionItem(actionItem); err != nil {
			return fmt.Errorf("error saving action item: %w", err)
		}
	}

	return nil
}

// handleExport exports the current recording to the specified format
func (m *Model) handleExport(format string) (tea.Model, tea.Cmd) {
	// Load full recording
	recording, err := m.db.GetRecording(m.currentRecordingID)
	if err != nil {
		m.err = fmt.Errorf("Failed to load recording for export: %w", err)
		return m, nil
	}

	// Generate filename
	timestamp := recording.CreatedAt.Format("20060102_150405")
	var filename string
	var exportErr error

	exporter := export.NewExporter()

	switch format {
	case "markdown":
		filename = fmt.Sprintf("data/exports/meeting_%s.md", timestamp)
		exportErr = exporter.ExportToMarkdown(recording, filename)
	case "json":
		filename = fmt.Sprintf("data/exports/meeting_%s.json", timestamp)
		exportErr = exporter.ExportToJSON(recording, filename)
	case "text":
		filename = fmt.Sprintf("data/exports/meeting_%s.txt", timestamp)
		exportErr = exporter.ExportToText(recording, filename)
	default:
		m.err = fmt.Errorf("Unknown export format: %s", format)
		return m, nil
	}

	if exportErr != nil {
		m.err = fmt.Errorf("Export failed: %w", exportErr)
		return m, nil
	}

	// Success - update processing message to show success
	m.err = nil
	m.processingMessage = fmt.Sprintf("✓ Exported successfully to: %s", filename)

	return m, nil
}

// View renders the current view
func (m *Model) View() string {
	// Show warning if audio is unavailable, but don't block the app
	var warningBanner string
	if m.err != nil {
		warningMsg := fmt.Sprintf("⚠  WARNING: %v", m.err)
		warningBanner = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Background(ColorGray).
			Padding(0, 1).
			Render(warningMsg) + "\n\n"
	}

	var content string
	switch m.view {
	case ViewMenu:
		content = m.viewMenu()
	case ViewRecording:
		content = m.viewRecording()
	case ViewProcessing:
		content = m.viewProcessing()
	case ViewAnalysis:
		content = m.viewAnalysis()
	case ViewHistory:
		content = m.viewHistory()
	case ViewSettings:
		content = m.viewSettings()
	default:
		content = "Unknown view"
	}

	return warningBanner + content
}

// Start starts the TUI application
func Start(db *storage.Database) error {
	p := tea.NewProgram(NewModel(db), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
