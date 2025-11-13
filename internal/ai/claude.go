package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// MeetingAnalysis contains the AI analysis results
type MeetingAnalysis struct {
	Summary     string
	KeyPoints   []string
	ActionItems []ActionItem
}

// ActionItem represents an actionable task
type ActionItem struct {
	Priority string // HIGH, MEDIUM, LOW
	Task     string
	Assignee string
}

// ClaudeClient wraps the Claude API client
type ClaudeClient struct {
	client *anthropic.Client
	model  string
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() (*ClaudeClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &ClaudeClient{
		client: client,
		model:  "claude-sonnet-4-5-20250929", // Latest Claude Sonnet
	}, nil
}

// AnalyzeMeeting analyzes a meeting transcript and generates insights
func (c *ClaudeClient) AnalyzeMeeting(ctx context.Context, transcript string) (*MeetingAnalysis, error) {
	prompt := c.buildAnalysisPrompt(transcript)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if textBlock := block.AsUnion().(*anthropic.ContentBlockText); textBlock != nil {
			responseText += textBlock.Text
		}
	}

	// Parse the structured response
	analysis, err := c.parseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return analysis, nil
}

// buildAnalysisPrompt creates the prompt for meeting analysis
func (c *ClaudeClient) buildAnalysisPrompt(transcript string) string {
	return fmt.Sprintf(`Analiza la siguiente transcripción de una reunión y proporciona:

1. RESUMEN: Un resumen conciso de 100-200 palabras que capture los temas principales y el propósito de la reunión.

2. PUNTOS CLAVE: 3-7 puntos clave más importantes discutidos, como lista con viñetas.

3. ACCIONABLES: Tareas o acciones identificadas, con:
   - Prioridad (ALTA/MEDIA/BAJA)
   - Descripción de la tarea
   - Persona asignada (si se menciona, sino usar "Sin asignar")

Formato de respuesta:

## RESUMEN
[Tu resumen aquí]

## PUNTOS CLAVE
- Punto 1
- Punto 2
- Punto 3

## ACCIONABLES
[ALTA] Descripción de la tarea - @Persona
[MEDIA] Descripción de la tarea - @Persona
[BAJA] Descripción de la tarea - @Persona

---

TRANSCRIPCIÓN:
%s`, transcript)
}

// parseAnalysisResponse parses Claude's response into structured data
func (c *ClaudeClient) parseAnalysisResponse(response string) (*MeetingAnalysis, error) {
	analysis := &MeetingAnalysis{
		KeyPoints:   []string{},
		ActionItems: []ActionItem{},
	}

	// Split response into sections
	sections := strings.Split(response, "##")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.Split(section, "\n")
		if len(lines) == 0 {
			continue
		}

		sectionTitle := strings.ToUpper(strings.TrimSpace(lines[0]))

		switch {
		case strings.Contains(sectionTitle, "RESUMEN") || strings.Contains(sectionTitle, "SUMMARY"):
			// Extract summary (everything after the title)
			summaryLines := lines[1:]
			analysis.Summary = strings.TrimSpace(strings.Join(summaryLines, "\n"))

		case strings.Contains(sectionTitle, "PUNTOS CLAVE") || strings.Contains(sectionTitle, "KEY POINTS"):
			// Extract bullet points
			for _, line := range lines[1:] {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Remove bullet markers
				line = strings.TrimPrefix(line, "-")
				line = strings.TrimPrefix(line, "•")
				line = strings.TrimPrefix(line, "*")
				line = strings.TrimSpace(line)
				if line != "" {
					analysis.KeyPoints = append(analysis.KeyPoints, line)
				}
			}

		case strings.Contains(sectionTitle, "ACCIONABLES") || strings.Contains(sectionTitle, "ACTION"):
			// Extract action items
			for _, line := range lines[1:] {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "---") {
					continue
				}

				item := c.parseActionItem(line)
				if item.Task != "" {
					analysis.ActionItems = append(analysis.ActionItems, item)
				}
			}
		}
	}

	return analysis, nil
}

// parseActionItem parses a single action item line
func (c *ClaudeClient) parseActionItem(line string) ActionItem {
	item := ActionItem{
		Priority: "MEDIA", // Default priority
		Assignee: "Sin asignar",
	}

	// Extract priority
	if strings.HasPrefix(line, "[") {
		endIdx := strings.Index(line, "]")
		if endIdx > 0 {
			priority := strings.ToUpper(strings.TrimSpace(line[1:endIdx]))
			item.Priority = priority
			line = strings.TrimSpace(line[endIdx+1:])
		}
	}

	// Extract assignee (after @ or - @)
	parts := strings.Split(line, "@")
	if len(parts) > 1 {
		item.Task = strings.TrimSpace(parts[0])
		// Remove trailing dash if present
		item.Task = strings.TrimSuffix(item.Task, "-")
		item.Task = strings.TrimSpace(item.Task)

		item.Assignee = "@" + strings.TrimSpace(parts[1])
	} else {
		item.Task = strings.TrimSpace(line)
	}

	return item
}

// GenerateSummary generates just a summary (lighter call)
func (c *ClaudeClient) GenerateSummary(ctx context.Context, transcript string) (string, error) {
	prompt := fmt.Sprintf(`Resume la siguiente transcripción de reunión en 2-3 oraciones concisas:

%s`, transcript)

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(500)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if textBlock := block.AsUnion().(*anthropic.ContentBlockText); textBlock != nil {
			responseText += textBlock.Text
		}
	}

	return strings.TrimSpace(responseText), nil
}
