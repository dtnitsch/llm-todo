package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// GetEnrichmentPath returns the path for the enrichment file for a session
func GetEnrichmentPath(sessionID string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	enrichmentDir := filepath.Join(home, ".llm-todo", "enrichment")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(enrichmentDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create enrichment directory: %w", err)
	}

	return filepath.Join(enrichmentDir, sessionID+".yaml"), nil
}

// GenerateEnrichmentFile creates a pre-filled YAML enrichment file for quick task creation
func GenerateEnrichmentFile(session *todo.Session, tasks []int64, taskTitles map[int64]string, filePath string) error {
	var b strings.Builder

	// Header with clear LLM instructions
	b.WriteString(fmt.Sprintf("# Auto-generated enrichment file for session: %s\n", session.ID))
	b.WriteString("# LLM: DO NOT EDIT IN PLACE - Re-write ENTIRE output and overwrite this file\n")
	b.WriteString(fmt.Sprintf("# Re-import after re-writing: llmtodo import %s\n\n", filePath))

	// Session metadata (optional)
	if session.Goal != "" {
		b.WriteString(fmt.Sprintf("goal: \"%s\"\n", escapeYAML(session.Goal)))
	} else {
		b.WriteString("goal: \"\"  # Optional: describe the purpose of this session\n")
	}

	if session.SuccessCriteria != "" {
		b.WriteString(fmt.Sprintf("success_criteria: \"%s\"\n", escapeYAML(session.SuccessCriteria)))
	}

	if session.Boundaries != "" && session.Boundaries != "[]" {
		b.WriteString(fmt.Sprintf("boundaries: \"%s\"\n", escapeYAML(session.Boundaries)))
	}

	if session.Deliverables != "" && session.Deliverables != "[]" {
		b.WriteString(fmt.Sprintf("deliverables: \"%s\"\n", escapeYAML(session.Deliverables)))
	}

	b.WriteString("\n")

	// Tasks section with example
	b.WriteString("tasks:\n")
	b.WriteString("  # EXAMPLE TASK - Shows all available fields - REMOVE THIS FROM YOUR OUTPUT\n")
	b.WriteString("  - id: example-task\n")
	b.WriteString("    title: \"Example: Implement feature X\"\n")
	b.WriteString("    priority: p0              # p0 (critical), p1 (important), p2 (normal), p3 (low), p4 (optional)\n")
	b.WriteString("    effort: m                 # xs, s, m (effort estimate)\n")
	b.WriteString("    type: task                # task, research, coordination, analysis, deliverable\n")
	b.WriteString("    files:\n")
	b.WriteString("      - \"path/to/file.go\"\n")
	b.WriteString("      - \"path/to/other.go\"\n")
	b.WriteString("    refs:                     # Reference URLs or docs\n")
	b.WriteString("      - \"https://docs.example.com\"\n")
	b.WriteString("    instructions:\n")
	b.WriteString("      must_do:\n")
	b.WriteString("        - \"Add validation\"\n")
	b.WriteString("        - \"Handle edge cases\"\n")
	b.WriteString("      must_not_do:\n")
	b.WriteString("        - \"Don't break existing API\"\n")
	b.WriteString("\n")
	b.WriteString("  # Your actual tasks (add fields as you know them)\n")

	// Actual tasks - MINIMAL (just id and title)
	for _, taskID := range tasks {
		title := taskTitles[taskID]
		b.WriteString(fmt.Sprintf("  - id: task-%d\n", taskID))
		b.WriteString(fmt.Sprintf("    title: \"%s\"\n", escapeYAML(title)))
		b.WriteString("\n")
	}

	// Write to file
	if err := os.WriteFile(filePath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("failed to write enrichment file: %w", err)
	}

	return nil
}

// escapeYAML escapes double quotes in YAML strings
func escapeYAML(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}
