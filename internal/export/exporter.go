package export

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/storo/scribescli/pkg/models"
)

// Exporter handles exporting recordings to different formats
type Exporter struct{}

// NewExporter creates a new exporter
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportToMarkdown exports a recording to Markdown format
func (e *Exporter) ExportToMarkdown(r *models.Recording, filename string) error {
	var md strings.Builder

	// Header
	md.WriteString(fmt.Sprintf("# %s\n\n", r.Title))
	md.WriteString(fmt.Sprintf("**Fecha:** %s\n\n", r.CreatedAt.Format("2006-01-02 15:04")))
	md.WriteString(fmt.Sprintf("**Duración:** %s\n\n", formatDuration(time.Duration(r.Duration)*time.Second)))

	if r.Language != "" {
		md.WriteString(fmt.Sprintf("**Idioma:** %s\n\n", r.Language))
	}

	md.WriteString("---\n\n")

	// Summary
	if r.Summary != "" {
		md.WriteString("## 📋 Resumen\n\n")
		md.WriteString(r.Summary)
		md.WriteString("\n\n")
	}

	// Key Points
	if len(r.KeyPoints) > 0 {
		md.WriteString("## ⭐ Puntos Clave\n\n")
		for _, point := range r.KeyPoints {
			md.WriteString(fmt.Sprintf("- %s\n", point))
		}
		md.WriteString("\n")
	}

	// Action Items
	if len(r.ActionItems) > 0 {
		md.WriteString("## ✓ Accionables\n\n")
		for _, item := range r.ActionItems {
			priority := getPriorityEmoji(item.Priority)
			status := "[ ]"
			if item.Completed {
				status = "[x]"
			}
			md.WriteString(fmt.Sprintf("- %s **%s** %s %s - %s\n",
				status, priority, item.Priority, item.Task, item.Assignee))
		}
		md.WriteString("\n")
	}

	// Transcript
	if r.Transcript != "" {
		md.WriteString("## 📝 Transcripción\n\n")
		md.WriteString("```\n")
		md.WriteString(r.Transcript)
		md.WriteString("\n```\n\n")
	}

	// Footer
	md.WriteString("---\n\n")
	md.WriteString("*Generado por ScribesAI - Meeting Intelligence System*\n")
	md.WriteString(fmt.Sprintf("*Exportado: %s*\n", time.Now().Format("2006-01-02 15:04")))

	// Write to file
	return os.WriteFile(filename, []byte(md.String()), 0644)
}

// ExportToJSON exports a recording to JSON format
func (e *Exporter) ExportToJSON(r *models.Recording, filename string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// ExportToText exports a recording to plain text format
func (e *Exporter) ExportToText(r *models.Recording, filename string) error {
	var txt strings.Builder

	// Header
	txt.WriteString(strings.Repeat("=", 80))
	txt.WriteString("\n")
	txt.WriteString(fmt.Sprintf("  %s\n", r.Title))
	txt.WriteString(strings.Repeat("=", 80))
	txt.WriteString("\n\n")

	txt.WriteString(fmt.Sprintf("Fecha: %s\n", r.CreatedAt.Format("2006-01-02 15:04")))
	txt.WriteString(fmt.Sprintf("Duración: %s\n", formatDuration(time.Duration(r.Duration)*time.Second)))

	if r.Language != "" {
		txt.WriteString(fmt.Sprintf("Idioma: %s\n", r.Language))
	}

	txt.WriteString("\n")
	txt.WriteString(strings.Repeat("-", 80))
	txt.WriteString("\n\n")

	// Summary
	if r.Summary != "" {
		txt.WriteString("RESUMEN\n")
		txt.WriteString(strings.Repeat("-", 80))
		txt.WriteString("\n")
		txt.WriteString(wrapText(r.Summary, 80))
		txt.WriteString("\n\n")
	}

	// Key Points
	if len(r.KeyPoints) > 0 {
		txt.WriteString("PUNTOS CLAVE\n")
		txt.WriteString(strings.Repeat("-", 80))
		txt.WriteString("\n")
		for i, point := range r.KeyPoints {
			txt.WriteString(fmt.Sprintf("%d. %s\n", i+1, point))
		}
		txt.WriteString("\n")
	}

	// Action Items
	if len(r.ActionItems) > 0 {
		txt.WriteString("ACCIONABLES\n")
		txt.WriteString(strings.Repeat("-", 80))
		txt.WriteString("\n")
		for i, item := range r.ActionItems {
			status := "[ ]"
			if item.Completed {
				status = "[X]"
			}
			txt.WriteString(fmt.Sprintf("%d. %s [%s] %s - %s\n",
				i+1, status, item.Priority, item.Task, item.Assignee))
		}
		txt.WriteString("\n")
	}

	// Transcript
	if r.Transcript != "" {
		txt.WriteString("TRANSCRIPCIÓN\n")
		txt.WriteString(strings.Repeat("-", 80))
		txt.WriteString("\n")
		txt.WriteString(r.Transcript)
		txt.WriteString("\n\n")
	}

	// Footer
	txt.WriteString(strings.Repeat("=", 80))
	txt.WriteString("\n")
	txt.WriteString("Generado por ScribesAI - Meeting Intelligence System\n")
	txt.WriteString(fmt.Sprintf("Exportado: %s\n", time.Now().Format("2006-01-02 15:04")))
	txt.WriteString(strings.Repeat("=", 80))
	txt.WriteString("\n")

	// Write to file
	return os.WriteFile(filename, []byte(txt.String()), 0644)
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
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func getPriorityEmoji(priority string) string {
	switch strings.ToUpper(priority) {
	case "HIGH", "ALTA":
		return "🔴"
	case "MEDIUM", "MEDIA":
		return "🟡"
	case "LOW", "BAJA":
		return "🟢"
	default:
		return "⚪"
	}
}

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
