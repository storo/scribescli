package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// viewMenu renders the main menu
func (m *Model) viewMenu() string {
	var b strings.Builder

	// Header
	header := RenderHeader("S C R I B E S  A I", "Meeting Intelligence System v1.0")
	b.WriteString(header)
	b.WriteString("\n\n")

	// HAL Eye with message
	message := GetHALQuote("idle")
	eye := m.halEye.RenderWithMessage(message)
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(eye))
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.menuItems {
		cursor := "  "
		if i == m.menuCursor {
			cursor = "► "
			b.WriteString(SelectedMenuItemStyle.Render(cursor + item))
		} else {
			b.WriteString(MenuItemStyle.Render(cursor + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	// Help text
	help := HelpStyle.Render("↑/↓: Navigate  Enter: Select  Q: Quit")
	b.WriteString(help)

	// Center everything
	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewRecording renders the recording view
func (m *Model) viewRecording() string {
	var b strings.Builder

	// Ensure minimum width
	width := ensureMinWidth(m.width, 70)

	// Top border
	topBorder := safeRepeat("═", width-2)
	b.WriteString("╔" + topBorder + "╗\n")

	// Title bar with status
	duration := formatDuration(m.recordingTime)
	status := "RECORDING"
	if m.paused {
		status = "PAUSED"
	}

	titleLeft := "  S C R I B E S  A I"
	titleRight := fmt.Sprintf("[●] %s  [■] STOP     ", duration)
	middleSpace := width - 2 - len(titleLeft) - len(titleRight)
	if middleSpace < 1 {
		middleSpace = 1
	}
	titleBar := titleLeft + safeRepeat(" ", middleSpace) + titleRight
	b.WriteString("║" + titleBar + "║\n")

	// Separator
	b.WriteString("╠" + topBorder + "╣\n")

	// Main content area
	b.WriteString("║" + safeRepeat(" ", width-2) + "║\n")

	// HAL Eye (compact version)
	eyeLines := strings.Split(m.halEye.RenderCompact(), "\n")
	for _, line := range eyeLines {
		lineWidth := lipgloss.Width(line)
		padding := (width - lineWidth) / 2
		if padding < 0 {
			padding = 0
		}
		rightPadding := safePadding(width-2, padding+lineWidth)
		b.WriteString("║" + safeRepeat(" ", padding) + line +
			safeRepeat(" ", rightPadding) + "║\n")
	}

	// Status message
	var statusMsg string
	if m.paused {
		statusMsg = "⏸  PAUSED"
	} else {
		statusMsg = fmt.Sprintf("🎤 %s", status)
	}

	quote := GetHALQuote("recording")
	statusLine := fmt.Sprintf("%s     \"%s\"", statusMsg, quote)
	statusLineLen := len(statusLine)
	padding := (width - statusLineLen) / 2
	if padding < 0 {
		padding = 0
	}
	rightPadding := safePadding(width-2, padding+statusLineLen)
	b.WriteString("║" + safeRepeat(" ", padding) + statusLine +
		safeRepeat(" ", rightPadding) + "║\n")

	b.WriteString("║" + safeRepeat(" ", width-2) + "║\n")

	// Transcript box
	transcriptTitle := "  LIVE TRANSCRIPT:"
	titlePadding := safePadding(width-4, len(transcriptTitle))
	b.WriteString("║  " + transcriptTitle + safeRepeat(" ", titlePadding) + "║\n")

	transcriptBoxWidth := width - 8
	if transcriptBoxWidth < 10 {
		transcriptBoxWidth = 10
	}
	transcriptBox := "  ┌" + safeRepeat("─", transcriptBoxWidth) + "┐"
	transcriptPadding := safePadding(width-2, len(transcriptBox))
	b.WriteString("║" + transcriptBox + safeRepeat(" ", transcriptPadding) + "║\n")

	// Show last 3 transcript lines
	transcriptLines := m.transcript
	if len(transcriptLines) > 3 {
		transcriptLines = transcriptLines[len(transcriptLines)-3:]
	}

	maxLineWidth := width - 10
	if maxLineWidth < 10 {
		maxLineWidth = 10
	}

	for i := 0; i < 3; i++ {
		var line string
		if i < len(transcriptLines) {
			line = transcriptLines[i]
			if len(line) > maxLineWidth {
				line = line[:maxLineWidth-3] + "..."
			}
		}
		linePadding := safePadding(transcriptBoxWidth-2, len(line))
		fullLine := "║  │ " + line + safeRepeat(" ", linePadding) + "│"
		finalPadding := safePadding(width-2, len(fullLine))
		b.WriteString(fullLine + safeRepeat(" ", finalPadding) + "║\n")
	}

	transcriptBoxEnd := "  └" + safeRepeat("─", transcriptBoxWidth) + "┘"
	endPadding := safePadding(width-2, len(transcriptBoxEnd))
	b.WriteString("║" + transcriptBoxEnd + safeRepeat(" ", endPadding) + "║\n")

	b.WriteString("║" + safeRepeat(" ", width-2) + "║\n")

	// Audio level bar
	levelLabel := "  AUDIO LEVEL: "
	percentage := fmt.Sprintf(" %3.0f%%", m.audioLevel*100)
	fixedWidth := len(levelLabel) + len(percentage) + 2 // 2 for borders
	barWidth := width - fixedWidth
	if barWidth < 10 {
		barWidth = 10
	}
	levelBar := renderSimpleProgressBar(m.audioLevel, barWidth)
	audioLevelLine := levelLabel + levelBar + percentage
	audioPadding := safePadding(width-2, len(audioLevelLine))
	b.WriteString("║" + audioLevelLine + safeRepeat(" ", audioPadding) + "║\n")

	// Status line
	var statusText string
	if m.paused {
		statusText = "  STATUS: Recording paused"
	} else {
		statusText = "  STATUS: Transcribing in real-time..."
	}
	statusPadding := safePadding(width-2, len(statusText))
	b.WriteString("║" + statusText + safeRepeat(" ", statusPadding) + "║\n")

	b.WriteString("║" + safeRepeat(" ", width-2) + "║\n")

	// Bottom border
	b.WriteString("╠" + topBorder + "╣\n")

	// Help bar
	help := "  [R]ecord  [S]top  [P]ause  [H]istory  [Q]uit"
	helpPadding := safePadding(width-2, len(help))
	b.WriteString("║" + help + safeRepeat(" ", helpPadding) + "║\n")
	b.WriteString("╚" + topBorder + "╝\n")

	return b.String()
}

// viewAnalysis renders the analysis results view
func (m *Model) viewAnalysis() string {
	var b strings.Builder

	// Header
	duration := formatDuration(m.recordingTime)
	header := fmt.Sprintf("ANALYSIS COMPLETE                              Duration: %s", duration)
	b.WriteString(DoubleBorderStyle.Render(header))
	b.WriteString("\n\n")

	// Summary section
	b.WriteString(SectionHeaderStyle.Render("📋 RESUMEN"))
	b.WriteString("\n")
	summaryBox := BorderStyle.Width(m.width - 10).Render(m.summary)
	b.WriteString(summaryBox)
	b.WriteString("\n\n")

	// Key points section
	b.WriteString(SectionHeaderStyle.Render("⭐ PUNTOS CLAVE"))
	b.WriteString("\n")
	for _, point := range m.keyPoints {
		b.WriteString(BulletStyle.Render("• " + point))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Action items section
	b.WriteString(SectionHeaderStyle.Render("✓ ACCIONABLES"))
	b.WriteString("\n")
	for _, item := range m.actionItems {
		priority := RenderPriorityBadge(item.Priority)
		actionLine := fmt.Sprintf("%s  %s - %s", priority, item.Task, item.Assignee)
		b.WriteString(BulletStyle.Render(actionLine))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	help := HelpStyle.Render("[E]xport  [T]ranscript  [B]ack  [Q]uit")
	b.WriteString(help)

	return b.String()
}

// viewHistory renders the history view
func (m *Model) viewHistory() string {
	var b strings.Builder

	header := RenderHeader("RECORDING HISTORY", "Previous meetings and transcriptions")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Placeholder for history
	b.WriteString(lipgloss.NewStyle().Foreground(ColorWhite).Render("No recordings yet. Start a new recording from the main menu."))
	b.WriteString("\n\n")

	help := HelpStyle.Render("[B]ack  [Q]uit")
	b.WriteString(help)

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewSettings renders the settings view
func (m *Model) viewSettings() string {
	var b strings.Builder

	header := RenderHeader("SETTINGS", "Configure ScribesAI")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Settings options
	settings := []string{
		"API Key: " + maskAPIKey(),
		"Sample Rate: 16000 Hz",
		"Channels: Mono",
		"Model: Vosk Multi-language",
	}

	for _, setting := range settings {
		b.WriteString(MenuItemStyle.Render("  " + setting))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	help := HelpStyle.Render("[B]ack  [Q]uit")
	b.WriteString(help)

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// Helper functions

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func renderSimpleProgressBar(level float32, width int) string {
	filled := int(level * float32(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return bar
}

func maskAPIKey() string {
	// TODO: Load from config
	return "sk-ant-*********************"
}
