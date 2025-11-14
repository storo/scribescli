package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/storo/scribescli/internal/audio"
	"github.com/storo/scribescli/internal/storage"
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
	halEye   *HALEye
	recorder *audio.Recorder
	db       *storage.Database

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

	// Window size
	width  int
	height int

	// Error message
	err error

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
	// This prevents PortAudio crashes in WSL2 at startup

	return &Model{
		view:     ViewMenu,
		halEye:   NewHALEye(),
		recorder: nil, // Initialize lazily
		db:       db,
		err:      nil,
		menuItems: []string{
			"New Recording",
			"History",
			"Settings",
			"Quit",
		},
		keys: DefaultKeyMap(),
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

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.recorder != nil {
				m.recorder.Close()
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
					m.err = fmt.Errorf("Cannot initialize audio: %v\n\nWSL2 users: Run './scripts/setup-wsl-audio.sh'", err)
					return m, nil
				}
				m.recorder = recorder
			}

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
			m.recorder.Stop()
			m.recording = false
			m.halEye.SetState("processing")

			// Save recording
			timestamp := time.Now().Format("20060102_150405")
			filename := fmt.Sprintf("data/recording_%s.wav", timestamp)
			if err := m.recorder.SaveRecording(filename); err != nil {
				m.err = err
			}

			// TODO: Start transcription and AI analysis
			// For now, go to analysis view
			m.view = ViewAnalysis
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
	// TODO: Handle analysis view keys (export, etc.)
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
