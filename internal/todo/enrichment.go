package todo

import (
	"encoding/json"
	"fmt"
)

// EnrichmentType represents different types of task enrichment
type EnrichmentType string

const (
	EnrichmentInstructions EnrichmentType = "instructions"  // must_do/must_not_do
	EnrichmentFiles        EnrichmentType = "files"         // related file paths
	EnrichmentOutput       EnrichmentType = "output"        // expected deliverable
	EnrichmentContext      EnrichmentType = "context"       // notes/background
	EnrichmentDependencies EnrichmentType = "dependencies"  // prerequisite tasks
)

// EnrichmentStatus tracks what enrichments exist on a task (0-5 scale)
type EnrichmentStatus struct {
	HasInstructions  bool `json:"has_instructions"`
	HasFiles         bool `json:"has_files"`
	HasOutput        bool `json:"has_output"`
	HasContext       bool `json:"has_context"`
	HasDependencies  bool `json:"has_dependencies"`
}

// EnrichmentScore calculates completeness (0-5)
func (e *EnrichmentStatus) Score() int {
	score := 0
	if e.HasInstructions {
		score++
	}
	if e.HasFiles {
		score++
	}
	if e.HasOutput {
		score++
	}
	if e.HasContext {
		score++
	}
	if e.HasDependencies {
		score++
	}
	return score
}

// Missing returns list of missing enrichment types
func (e *EnrichmentStatus) Missing() []EnrichmentType {
	var missing []EnrichmentType
	if !e.HasInstructions {
		missing = append(missing, EnrichmentInstructions)
	}
	if !e.HasFiles {
		missing = append(missing, EnrichmentFiles)
	}
	if !e.HasOutput {
		missing = append(missing, EnrichmentOutput)
	}
	if !e.HasContext {
		missing = append(missing, EnrichmentContext)
	}
	if !e.HasDependencies {
		missing = append(missing, EnrichmentDependencies)
	}
	return missing
}

// GetEnrichmentStatus calculates enrichment status for a task
func GetEnrichmentStatus(task *Task) *EnrichmentStatus {
	status := &EnrichmentStatus{}

	// Check instructions (non-empty JSON with must_do or must_not_do)
	if task.Instructions != "" && task.Instructions != "{}" {
		var instructions map[string][]string
		if json.Unmarshal([]byte(task.Instructions), &instructions) == nil {
			if len(instructions["must_do"]) > 0 || len(instructions["must_not_do"]) > 0 {
				status.HasInstructions = true
			}
		}
	}

	// Check files
	if task.Files != "" && task.Files != "[]" {
		var files []string
		if json.Unmarshal([]byte(task.Files), &files) == nil && len(files) > 0 {
			status.HasFiles = true
		}
	}

	// Check output
	if task.Output != "" {
		status.HasOutput = true
	}

	// Check context (notes)
	if task.Notes != "" {
		status.HasContext = true
	}

	// Check dependencies
	if task.DependantIDs != "" && task.DependantIDs != "[]" {
		var deps []int
		if json.Unmarshal([]byte(task.DependantIDs), &deps) == nil && len(deps) > 0 {
			status.HasDependencies = true
		}
	}

	return status
}

// GetEnrichmentHint returns a hint for enriching a task
func GetEnrichmentHint(task *Task) string {
	status := GetEnrichmentStatus(task)
	score := status.Score()

	if score == 5 {
		return "" // Fully enriched
	}

	missing := status.Missing()
	if len(missing) == 0 {
		return ""
	}

	// Build hint command
	hint := fmt.Sprintf("Task has %d/5 enrichments. Add: llmtodo enrich %d", score, task.ID)

	// Suggest flags based on what's missing
	examples := []string{}
	for _, m := range missing {
		switch m {
		case EnrichmentInstructions:
			examples = append(examples, "--must-do 'Concise action item'")
		case EnrichmentFiles:
			examples = append(examples, "--files 'path/to/file.go'")
		case EnrichmentOutput:
			examples = append(examples, "--output 'Concrete deliverable'")
		case EnrichmentContext:
			examples = append(examples, "--notes 'Why this task exists'")
		case EnrichmentDependencies:
			examples = append(examples, "--deps '1,2'")
		}
	}

	if len(examples) > 0 {
		hint += " " + examples[0]
		if len(examples) > 1 {
			hint += " ..."
		}
	}

	return hint
}

// EnrichmentSuggestion represents a specific enrichment recommendation
type EnrichmentSuggestion struct {
	Type        EnrichmentType `json:"type"`
	Description string         `json:"description"`
	Example     string         `json:"example,omitempty"`
}

// GetEnrichmentSuggestions returns detailed suggestions for enriching a task
func GetEnrichmentSuggestions(task *Task) []EnrichmentSuggestion {
	status := GetEnrichmentStatus(task)
	missing := status.Missing()

	var suggestions []EnrichmentSuggestion
	for _, m := range missing {
		var suggestion EnrichmentSuggestion
		suggestion.Type = m

		switch m {
		case EnrichmentInstructions:
			suggestion.Description = "Add concise must_do/must_not_do (imperative, no prose)"
			suggestion.Example = fmt.Sprintf("llmtodo enrich %d --must-do 'Add validation' --must-not 'Skip error handling'", task.ID)
		case EnrichmentFiles:
			suggestion.Description = "Specify file paths (no descriptions)"
			suggestion.Example = fmt.Sprintf("llmtodo enrich %d --files 'main.go,test.go'", task.ID)
		case EnrichmentOutput:
			suggestion.Description = "Define concrete deliverable (single sentence)"
			suggestion.Example = fmt.Sprintf("llmtodo enrich %d --output 'Working API endpoint with tests'", task.ID)
		case EnrichmentContext:
			suggestion.Description = "Add WHY (1-2 sentences, present tense, no implementation details)"
			suggestion.Example = fmt.Sprintf("llmtodo enrich %d --notes 'Validates enrichment works for cold-start LLMs'", task.ID)
		case EnrichmentDependencies:
			suggestion.Description = "Specify prerequisite task IDs"
			suggestion.Example = fmt.Sprintf("llmtodo enrich %d --deps '1,2'", task.ID)
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}
