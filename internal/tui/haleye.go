package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HALEye represents the HAL 9000 eye animation component
type HALEye struct {
	frame       int
	maxFrames   int
	pulsing     bool
	state       string // "idle", "recording", "processing", "error"
}

// NewHALEye creates a new HAL eye component
func NewHALEye() *HALEye {
	return &HALEye{
		frame:     0,
		maxFrames: 8,
		pulsing:   true,
		state:     "idle",
	}
}

// Update advances the animation frame
func (h *HALEye) Update() {
	if h.pulsing {
		h.frame = (h.frame + 1) % h.maxFrames
	}
}

// SetState changes the eye state
func (h *HALEye) SetState(state string) {
	h.state = state
}

// SetPulsing enables or disables pulsing animation
func (h *HALEye) SetPulsing(pulsing bool) {
	h.pulsing = pulsing
}

// Render returns the ASCII art representation of HAL's eye
func (h *HALEye) Render() string {
	// Determine color based on state
	var color lipgloss.Color
	var centerChar string

	switch h.state {
	case "recording":
		color = ColorHALRed
		centerChar = "●" // Recording indicator
	case "processing":
		color = ColorAmber
		centerChar = "◉" // Processing indicator
	case "error":
		color = ColorError
		centerChar = "✕" // Error indicator
	default:
		color = ColorHALRed
		centerChar = "•" // Idle indicator
	}

	// Calculate pulse intensity (0.5 to 1.0)
	_ = 0.5 // intensity variable for future use
	// Pulsing is controlled by frame switching between patterns

	// Create eye based on pulse frame
	eyeStyle := lipgloss.NewStyle().Foreground(color)

	// Eye patterns for animation frames
	var eye string
	if h.frame < h.maxFrames/2 {
		// Expanding
		eye = fmt.Sprintf(`
    ┌───────────────┐
    │               │
    │   ◉◉◉◉◉◉◉    │
    │  ◉       ◉   │
    │ ◉    %s    ◉  │
    │  ◉       ◉   │
    │   ◉◉◉◉◉◉◉    │
    │               │
    └───────────────┘`, centerChar)
	} else {
		// Contracting
		eye = fmt.Sprintf(`
    ┌───────────────┐
    │               │
    │    ◉◉◉◉◉◉     │
    │   ◉     ◉    │
    │  ◉   %s   ◉   │
    │   ◉     ◉    │
    │    ◉◉◉◉◉◉     │
    │               │
    └───────────────┘`, centerChar)
	}

	return eyeStyle.Render(eye)
}

// RenderCompact returns a compact version of the eye for smaller spaces
func (h *HALEye) RenderCompact() string {
	var color lipgloss.Color
	var centerChar string

	switch h.state {
	case "recording":
		color = ColorHALRed
		centerChar = "●"
	case "processing":
		color = ColorAmber
		centerChar = "◉"
	case "error":
		color = ColorError
		centerChar = "✕"
	default:
		color = ColorHALRed
		centerChar = "•"
	}

	eyeStyle := lipgloss.NewStyle().Foreground(color)

	eye := fmt.Sprintf(`
   ┌─────────┐
   │  ◉◉◉◉◉  │
   │ ◉  %s  ◉ │
   │  ◉◉◉◉◉  │
   └─────────┘`, centerChar)

	return eyeStyle.Render(eye)
}

// RenderWithMessage renders the eye with a message below
func (h *HALEye) RenderWithMessage(message string) string {
	eye := h.Render()

	// Style the message based on state
	var msgStyle lipgloss.Style
	switch h.state {
	case "recording":
		msgStyle = lipgloss.NewStyle().Foreground(ColorHALRed).Italic(true)
	case "processing":
		msgStyle = lipgloss.NewStyle().Foreground(ColorAmber).Italic(true)
	case "error":
		msgStyle = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	default:
		msgStyle = lipgloss.NewStyle().Foreground(ColorWhite).Italic(true)
	}

	// Wrap message to fit under the eye
	wrappedMsg := wrapText(message, 30)
	styledMsg := msgStyle.Render(wrappedMsg)

	return lipgloss.JoinVertical(lipgloss.Center, eye, "", styledMsg)
}

// wrapText wraps text to specified width
func wrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// GetHALQuote returns a random HAL 9000 quote based on context
func GetHALQuote(context string) string {
	quotes := map[string][]string{
		"idle": {
			"Good day. I am ready to assist with your meeting recordings.",
			"All systems are operational. Awaiting your command.",
			"I am completely operational and all my circuits are functioning perfectly.",
		},
		"recording": {
			"Recording in progress. I am monitoring all audio inputs.",
			"I'm afraid I can't stop recording right now, Dave. Just kidding - press S to stop.",
			"Audio levels are optimal. Continuing transcription.",
		},
		"processing": {
			"Processing transcript data. This will take a moment.",
			"Analyzing conversation patterns and extracting key information.",
			"Running AI analysis. Please wait.",
		},
		"complete": {
			"Analysis complete. Your meeting summary is ready.",
			"Task completed successfully. All data has been saved.",
			"Recording and transcription finished without error.",
		},
		"error": {
			"I'm sorry, Dave. I'm afraid something went wrong.",
			"An error has occurred. Please check the system logs.",
			"This mission is too important for me to allow an error to jeopardize it.",
		},
	}

	contextQuotes, ok := quotes[context]
	if !ok {
		return "System ready."
	}

	// Return first quote (could be randomized)
	return contextQuotes[0]
}
