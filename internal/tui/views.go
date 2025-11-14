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

// viewProcessing renders the processing/analysis view
func (m *Model) viewProcessing() string {
	var b strings.Builder

	// Header
	header := RenderHeader("A N A L Y Z I N G", "Claude AI is analyzing your meeting")
	b.WriteString(header)
	b.WriteString("\n\n")

	// HAL Eye with processing animation
	message := GetHALQuote("processing")
	eye := m.halEye.RenderWithMessage(message)
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(eye))
	b.WriteString("\n\n")

	// Processing message
	statusMsg := lipgloss.NewStyle().
		Foreground(ColorAmber).
		Bold(true).
		Render(m.processingMessage)
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(statusMsg))
	b.WriteString("\n\n")

	// Spinner or progress indicator
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	spinnerChar := string(spinner[m.halEye.frame%len(spinner)])
	spinnerLine := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Render(fmt.Sprintf("%s Extracting key points...", spinnerChar))
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(spinnerLine))
	b.WriteString("\n")

	spinnerLine2 := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Render(fmt.Sprintf("%s Identifying action items...", spinnerChar))
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(spinnerLine2))
	b.WriteString("\n")

	spinnerLine3 := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Render(fmt.Sprintf("%s Generating summary...", spinnerChar))
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(spinnerLine3))
	b.WriteString("\n\n")

	// Help text
	help := HelpStyle.Render("Please wait... This may take 10-30 seconds")
	b.WriteString(help)

	// Center everything
	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// viewAnalysis renders the analysis results view
func (m *Model) viewAnalysis() string {
	var b strings.Builder

	// Ensure minimum width for analysis display
	width := ensureMinWidth(m.width, 80)

	// Header with duration
	duration := formatDuration(m.recordingTime)
	headerText := fmt.Sprintf("A N A L Y S I S   C O M P L E T E                    Duration: %s", duration)
	header := TitleStyle.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Compact HAL eye
	eye := m.halEye.RenderCompact()
	quote := GetHALQuote("complete")
	eyeWithQuote := lipgloss.JoinVertical(lipgloss.Center, eye, "",
		lipgloss.NewStyle().Foreground(ColorWhite).Italic(true).Render(quote))
	b.WriteString(lipgloss.NewStyle().MarginLeft(2).Render(eyeWithQuote))
	b.WriteString("\n\n")

	// Summary section
	b.WriteString(SectionHeaderStyle.Render("📋 RESUMEN"))
	b.WriteString("\n")
	if m.summary != "" {
		summaryBox := BorderStyle.Width(width - 10).Render(m.summary)
		b.WriteString(summaryBox)
	} else {
		b.WriteString(HelpStyle.Render("  No summary available"))
	}
	b.WriteString("\n\n")

	// Key points section
	b.WriteString(SectionHeaderStyle.Render("⭐ PUNTOS CLAVE"))
	b.WriteString("\n")
	if len(m.keyPoints) > 0 {
		for i, point := range m.keyPoints {
			pointNum := lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true).
				Render(fmt.Sprintf("%d.", i+1))
			b.WriteString(BulletStyle.Render(pointNum + " " + point))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(HelpStyle.Render("  No key points extracted"))
	}
	b.WriteString("\n")

	// Action items section
	b.WriteString(SectionHeaderStyle.Render("✓ ACCIONABLES"))
	b.WriteString("\n")
	if len(m.actionItems) > 0 {
		for i, item := range m.actionItems {
			priority := RenderPriorityBadge(item.Priority)
			actionLine := fmt.Sprintf("%d. %s  %s - %s",
				i+1, priority, item.Task, item.Assignee)
			b.WriteString(BulletStyle.Render(actionLine))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(HelpStyle.Render("  No action items identified"))
	}
	b.WriteString("\n")

	// Export options
	b.WriteString(SectionHeaderStyle.Render("💾 EXPORT OPTIONS"))
	b.WriteString("\n")
	exportHelp := []string{
		"  [M] Export to Markdown",
		"  [J] Export to JSON",
		"  [T] Export to Text",
	}
	for _, opt := range exportHelp {
		b.WriteString(MenuItemStyle.Render(opt))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	help := HelpStyle.Render("[M]arkdown  [J]SON  [T]ext  [B]ack  [Q]uit")
	b.WriteString(help)

	return b.String()
}

// viewHistory renders the history view
func (m *Model) viewHistory() string {
	var b strings.Builder

	header := RenderHeader("RECORDING HISTORY", "Previous meetings and transcriptions")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Load recordings if not cached
	if m.recordings == nil {
		recordings, err := m.db.ListRecordings(m.historyLimit, m.historyOffset)
		if err != nil {
			b.WriteString(lipgloss.NewStyle().Foreground(ColorRed).Render(fmt.Sprintf("Error loading recordings: %v", err)))
			b.WriteString("\n\n")
			help := HelpStyle.Render("[B]ack  [Q]uit")
			b.WriteString(help)
			content := b.String()
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
		}

		if len(recordings) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(ColorWhite).Render("No recordings yet."))
			b.WriteString("\n\n")
			b.WriteString(HelpStyle.Render("[N]ew Recording  [B]ack  [Q]uit"))
			content := b.String()
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
		}

		m.recordings = recordings
	}

	// Display recordings
	for i, rec := range m.recordings {
		// Format date and duration
		date := rec.CreatedAt.Format("2006-01-02 15:04")
		duration := formatDuration(time.Duration(rec.Duration) * time.Second)

		// Highlight current selection
		var style lipgloss.Style
		prefix := "  "
		if i == m.menuCursor {
			style = SelectedMenuItemStyle
			prefix = "► "
		} else {
			style = MenuItemStyle
		}

		// Recording line
		line := fmt.Sprintf("%s[%d] %s - %s (%s)", prefix, rec.ID, rec.Title, date, duration)
		b.WriteString(style.Render(line))
		b.WriteString("\n")

		// Show summary preview if available
		if rec.Summary != "" && i == m.menuCursor {
			preview := rec.Summary
			if len(preview) > 80 {
				preview = preview[:77] + "..."
			}
			b.WriteString(HelpStyle.Render("    " + preview))
			b.WriteString("\n")
		}
	}

	// Pagination info
	b.WriteString("\n")
	totalPages := (m.historyOffset / m.historyLimit) + 1
	pageInfo := fmt.Sprintf("Page %d | Showing %d recordings", totalPages, len(m.recordings))
	b.WriteString(HelpStyle.Render(pageInfo))
	b.WriteString("\n\n")

	// Help text
	help := HelpStyle.Render("↑/↓: Navigate  Enter: View  [N]ext Page  [P]rev Page  [B]ack  [Q]uit")
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
