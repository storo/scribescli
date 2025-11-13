package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// HAL 9000 inspired color palette
var (
	// Primary colors
	ColorHALRed     = lipgloss.Color("#FF0000") // HAL's iconic red
	ColorAmber      = lipgloss.Color("#FFB000") // Warnings, active state
	ColorCRTGreen   = lipgloss.Color("#00FF41") // Main text (CRT terminal green)
	ColorCyan       = lipgloss.Color("#00FFFF") // Highlights
	ColorBackground = lipgloss.Color("#0A0E14") // Dark background
	ColorGray       = lipgloss.Color("#1A1F29") // Secondary background
	ColorWhite      = lipgloss.Color("#E6E6E6") // Secondary text

	// Status colors
	ColorSuccess = lipgloss.Color("#00FF41")
	ColorWarning = lipgloss.Color("#FFB000")
	ColorError   = lipgloss.Color("#FF0000")
	ColorInfo    = lipgloss.Color("#00FFFF")
)

// Styles for different components
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(ColorCRTGreen).
			Background(ColorBackground)

	// Title style
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorHALRed).
			Bold(true).
			Padding(0, 1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Italic(true)

	// Border styles
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(1, 2)

	DoubleBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.DoubleBorder()).
				BorderForeground(ColorHALRed).
				Padding(1, 2)

	// Menu item styles
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorCRTGreen).
			Padding(0, 2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorHALRed).
				Bold(true).
				Padding(0, 2)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Background(ColorGray).
			Padding(0, 1)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Italic(true).
			Padding(0, 1)

	// Recording indicator style
	RecordingStyle = lipgloss.NewStyle().
			Foreground(ColorHALRed).
			Bold(true).
			Blink(true)

	// Audio level bar style
	AudioLevelBarStyle = lipgloss.NewStyle().
				Foreground(ColorCRTGreen)

	// Transcript style
	TranscriptStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Background(ColorGray).
			Padding(1, 2).
			MarginTop(1)

	// Analysis section styles
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true).
				MarginTop(1)

	BulletStyle = lipgloss.NewStyle().
			Foreground(ColorCRTGreen).
			Padding(0, 1)

	// Priority badge styles
	HighPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorHALRed).
				Bold(true).
				Padding(0, 1)

	MediumPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorAmber).
				Bold(true).
				Padding(0, 1)

	LowPriorityStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorSuccess).
				Bold(true).
				Padding(0, 1)

	// Error message style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true).
			Padding(1, 2)
)

// RenderHeader renders the application header
func RenderHeader(title, subtitle string) string {
	t := TitleStyle.Render(title)
	s := SubtitleStyle.Render(subtitle)
	return lipgloss.JoinVertical(lipgloss.Center, t, s)
}

// RenderBox renders content in a bordered box
func RenderBox(content string, width int) string {
	style := BorderStyle.Width(width)
	return style.Render(content)
}

// RenderStatusBar renders the status bar with time and status
func RenderStatusBar(left, right string, width int) string {
	leftPart := StatusBarStyle.Render(left)
	rightPart := StatusBarStyle.Render(right)

	// Calculate spacing
	leftLen := lipgloss.Width(leftPart)
	rightLen := lipgloss.Width(rightPart)
	spacing := width - leftLen - rightLen
	if spacing < 0 {
		spacing = 0
	}

	spacer := lipgloss.NewStyle().Width(spacing).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPart, spacer, rightPart)
}

// RenderProgressBar renders an audio level progress bar
func RenderProgressBar(level float32, width int) string {
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

	percentage := int(level * 100)
	return AudioLevelBarStyle.Render(bar) + "  " + lipgloss.NewStyle().Foreground(ColorWhite).Render(fmt.Sprintf("%d%%", percentage))
}

// RenderPriorityBadge renders a priority badge
func RenderPriorityBadge(priority string) string {
	switch priority {
	case "HIGH", "ALTA":
		return HighPriorityStyle.Render("🔴 " + priority)
	case "MEDIUM", "MEDIA":
		return MediumPriorityStyle.Render("🟡 " + priority)
	case "LOW", "BAJA":
		return LowPriorityStyle.Render("🟢 " + priority)
	default:
		return priority
	}
}
