package importer

import (
	"fmt"
	"strings"
)

// ImportError represents a detailed import error with suggestions
type ImportError struct {
	TaskID      string
	Field       string
	Value       string
	Message     string
	Suggestion  string
	LineNumber  int
	AvailableIDs []int
}

func (e *ImportError) Error() string {
	var b strings.Builder

	b.WriteString("ERROR: Import failed")

	if e.TaskID != "" {
		b.WriteString(fmt.Sprintf(" for task %s", e.TaskID))
	}

	if e.Field != "" {
		b.WriteString(fmt.Sprintf(" (field: %s)", e.Field))
	}

	b.WriteString(fmt.Sprintf("\n   %s\n", e.Message))

	if e.Value != "" {
		b.WriteString(fmt.Sprintf("   Got: %q\n", e.Value))
	}

	if e.Suggestion != "" {
		b.WriteString(fmt.Sprintf("\n   HINT: %s\n", e.Suggestion))
	}

	if len(e.AvailableIDs) > 0 {
		b.WriteString(fmt.Sprintf("\n   Available task IDs: %v\n", e.AvailableIDs))
	}

	return b.String()
}

// ValidatePriority checks if priority is valid
func ValidatePriority(priority string) error {
	valid := map[string]bool{
		"p0": true,
		"p1": true,
		"p2": true,
		"p3": true,
		"p4": true,
	}

	if !valid[priority] {
		// Check for common typos
		suggestions := []string{}

		if strings.Contains(priority, "0") {
			suggestions = append(suggestions, "p0")
		}
		if strings.Contains(priority, "1") {
			suggestions = append(suggestions, "p1")
		}

		// Check for case issues
		lower := strings.ToLower(priority)
		if valid[lower] {
			suggestions = append(suggestions, lower)
		}

		suggestionMsg := ""
		if len(suggestions) > 0 {
			suggestionMsg = fmt.Sprintf("Did you mean: %s?", strings.Join(suggestions, " or "))
		} else {
			suggestionMsg = "Valid priorities: p0 (critical), p1 (important), p2 (normal), p3 (low), p4 (optional)"
		}

		return &ImportError{
			Field:      "priority",
			Value:      priority,
			Message:    "Invalid priority",
			Suggestion: suggestionMsg,
		}
	}

	return nil
}

// ValidateEffort checks if effort is valid
func ValidateEffort(effort string) error {
	valid := map[string]bool{
		"xs": true,
		"s":  true,
		"m":  true,
	}

	if !valid[effort] {
		suggestion := "Valid efforts: xs (extra small), s (small), m (medium)"

		// Check for common mistakes
		if strings.ToLower(effort) == "small" {
			suggestion = "Did you mean: s"
		} else if strings.ToLower(effort) == "medium" {
			suggestion = "Did you mean: m"
		} else if strings.Contains(strings.ToLower(effort), "extra") {
			suggestion = "Did you mean: xs"
		}

		return &ImportError{
			Field:      "effort",
			Value:      effort,
			Message:    "Invalid effort",
			Suggestion: suggestion,
		}
	}

	return nil
}

// ValidateTaskType checks if task type is valid
func ValidateTaskType(taskType string) error {
	valid := map[string]bool{
		"task":         true,
		"research":     true,
		"coordination": true,
		"analysis":     true,
		"deliverable":  true,
	}

	if !valid[taskType] {
		return &ImportError{
			Field:      "type",
			Value:      taskType,
			Message:    "Invalid task type",
			Suggestion: "Valid types: task, research, coordination, analysis, deliverable",
		}
	}

	return nil
}

// CheckForCommonTypos checks for common field name typos
func CheckForCommonTypos(data string) []string {
	typos := []string{}

	// YAML field-specific typos (check for "field:" pattern)
	// Note: Order matters - check longer strings before shorter ones to avoid false positives
	fieldTypos := []struct {
		typo    string
		correct string
	}{
		{"priortiy:", "priority:"},
		{"prioirty:", "priority:"},
		{"priorit:", "priority:"},
		{"insturctions:", "instructions:"},
		{"instrutions:", "instructions:"},
		{"instrucions:", "instructions:"},
		{"mustnt:", "must_not_do:"},
		{"mustnot:", "must_not_do:"},
		// Check for "must_not:" only when NOT followed by "_do"
		// (to avoid matching valid "must_not_do:")
	}

	lowerData := strings.ToLower(data)

	// Check for "must_not:" as a field (but not "must_not_do:")
	// Look for it as a complete word, not as part of "must_not_do:"
	lines := strings.Split(lowerData, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip if this line has the correct field name
		if strings.Contains(trimmed, "must_not_do:") {
			continue
		}
		// Now check if it has the typo
		if strings.Contains(trimmed, "must_not:") {
			typos = append(typos, "Found 'must_not' - did you mean 'must_not_do'?")
			break
		}
	}

	for _, pair := range fieldTypos {
		if strings.Contains(lowerData, pair.typo) {
			typos = append(typos, fmt.Sprintf("Found '%s' - did you mean '%s'?", strings.TrimSuffix(pair.typo, ":"), strings.TrimSuffix(pair.correct, ":")))
		}
	}

	return typos
}
