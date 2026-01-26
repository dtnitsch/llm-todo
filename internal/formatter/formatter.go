package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// FormatNext formats the next task with conditional sections
func FormatNext(output *todo.NextOutput, suggestions []todo.Suggestion) string {
	var b strings.Builder

	task := output.Task
	session := output.Session

	// Header with task ID
	b.WriteString(fmt.Sprintf("Task #%d: %s\n", task.ID, task.Task))

	// Metadata line (priority | effort | files)
	var metaParts []string
	if task.Priority != "" {
		metaParts = append(metaParts, task.Priority)
	}
	if task.Effort != "" {
		metaParts = append(metaParts, fmt.Sprintf("est: %s", task.Effort))
	}
	if task.Files != "" && task.Files != "[]" {
		var files []string
		if err := json.Unmarshal([]byte(task.Files), &files); err == nil && len(files) > 0 {
			metaParts = append(metaParts, strings.Join(files, ", "))
		}
	}
	if len(metaParts) > 0 {
		b.WriteString(strings.Join(metaParts, " | "))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// CONTEXT CHECK - warn if missing critical data
	hasFiles := task.Files != "" && task.Files != "[]"
	hasInstructions := task.Instructions != "" && task.Instructions != "{}"
	hasRefs := task.Refs != "" && task.Refs != "[]"

	missingContext := false
	if !hasFiles && !hasInstructions && !hasRefs && task.Notes == "" {
		missingContext = true
	}

	if missingContext {
		b.WriteString("MISSING CONTEXT: No files, instructions, or notes\n")
		b.WriteString("Hint: llmtodo enrich " + fmt.Sprintf("%d", task.ID) + "\n\n")
	}

	// SUGGESTIONS (if any)
	if len(suggestions) > 0 {
		b.WriteString("SUGGESTIONS:\n")
		for _, sug := range suggestions {
			b.WriteString(fmt.Sprintf("  - %s\n", sug.Message))
		}
		b.WriteString("\n")
	}

	// INSTRUCTIONS (must_do / must_not_do) - compact format
	hasInstructionsContent := false
	if task.Instructions != "" {
		var instructions map[string][]string
		if err := json.Unmarshal([]byte(task.Instructions), &instructions); err == nil {
			if mustDo := instructions["must_do"]; len(mustDo) > 0 {
				hasInstructionsContent = true
				for _, item := range mustDo {
					b.WriteString(fmt.Sprintf("  - %s\n", item))
				}
				b.WriteString("\n")
			}
			if mustNotDo := instructions["must_not_do"]; len(mustNotDo) > 0 {
				hasInstructionsContent = true
				b.WriteString("MUST NOT:\n")
				for _, item := range mustNotDo {
					b.WriteString(fmt.Sprintf("  ✗ %s\n", item))
				}
				b.WriteString("\n")
			}
		}
	}

	// Show session context as fallback when no instructions
	if !hasInstructionsContent && session != nil && session.Goal != "" {
		b.WriteString(fmt.Sprintf("Session: %s - %s\n", session.ID, session.Goal))
		b.WriteString("Session details: llmtodo session\n\n")
	}

	// REFS (research mode)
	if task.Refs != "" && task.Refs != "[]" {
		var refs []string
		if err := json.Unmarshal([]byte(task.Refs), &refs); err == nil && len(refs) > 0 {
			b.WriteString("REFERENCES:\n")
			for _, ref := range refs {
				b.WriteString(fmt.Sprintf("  - %s\n", ref))
			}
			b.WriteString("\n")
		}
	}

	// WAITING_ON (research mode)
	if task.WaitingOn != "" {
		b.WriteString(fmt.Sprintf("WAITING ON: %s\n\n", task.WaitingOn))
	}

	// OUTPUT/AUDIENCE (deliverable mode)
	if task.Output != "" {
		b.WriteString(fmt.Sprintf("OUTPUT: %s\n", task.Output))
		if task.Audience != "" {
			b.WriteString(fmt.Sprintf("AUDIENCE: %s\n", task.Audience))
		}
		b.WriteString("\n")
	}

	// NOTES
	if task.Notes != "" {
		b.WriteString(fmt.Sprintf("NOTES:\n%s\n\n", task.Notes))
	}

	// UPCOMING TASKS (show next 2-3 tasks)
	if len(output.UpcomingTasks) > 0 {
		b.WriteString("UPCOMING:\n")
		for _, upcoming := range output.UpcomingTasks {
			b.WriteString(fmt.Sprintf("  %d. %s", upcoming.ID, upcoming.Task))
			if upcoming.Priority != "" && upcoming.Priority != "p1" {
				b.WriteString(fmt.Sprintf(" (%s)", upcoming.Priority))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer - show specific command with task ID
	b.WriteString(fmt.Sprintf("llmtodo done %d\n", task.ID))

	return b.String()
}

// PrintFull prints full task details
func PrintFull(task *todo.Task) {
	fmt.Printf("Task #%d: %s\n", task.ID, task.Task)
	fmt.Printf("Status: %s\n", colorizeStatus(task.Status))
	fmt.Printf("Priority: %s (order: %d)\n", task.Priority, task.PriorityOrder)
	if task.Type != "task" {
		fmt.Printf("Type: %s\n", task.Type)
	}
	if task.Effort != "" {
		fmt.Printf("Effort: %s\n", task.Effort)
	}

	if task.Instructions != "" {
		fmt.Println("\nInstructions:")
		var instructions map[string][]string
		if err := json.Unmarshal([]byte(task.Instructions), &instructions); err == nil {
			if mustDo := instructions["must_do"]; len(mustDo) > 0 {
				fmt.Println("  Must do:")
				for _, item := range mustDo {
					fmt.Printf("    ✓ %s\n", item)
				}
			}
			if mustNotDo := instructions["must_not_do"]; len(mustNotDo) > 0 {
				fmt.Println("  Must NOT do:")
				for _, item := range mustNotDo {
					fmt.Printf("    ✗ %s\n", item)
				}
			}
		}
	}

	if task.Files != "" && task.Files != "[]" {
		var files []string
		if err := json.Unmarshal([]byte(task.Files), &files); err == nil && len(files) > 0 {
			fmt.Printf("\nFiles: %s\n", strings.Join(files, ", "))
		}
	}

	if task.Notes != "" {
		fmt.Printf("\nNotes:\n%s\n", task.Notes)
	}

	if task.DependantIDs != "" && task.DependantIDs != "[]" {
		var depIDs []int64
		if err := json.Unmarshal([]byte(task.DependantIDs), &depIDs); err == nil && len(depIDs) > 0 {
			fmt.Printf("\nDependencies: ")
			for i, id := range depIDs {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("#%d", id)
			}
			fmt.Println()
		}
	}

	if task.BlockingReason != "" {
		fmt.Printf("\nBLOCKED: %s\n", task.BlockingReason)
	}
}

// PrintStatus prints session status summary
func PrintStatus(session *todo.Session, stats map[string]int) {
	fmt.Printf("Session: %s (%s)\n", session.ID, session.Type)
	if session.Goal != "" {
		fmt.Printf("Goal: %s\n", session.Goal)
	}
	fmt.Printf("\nProgress:\n")
	fmt.Printf("  Total: %d\n", stats["total"])
	fmt.Printf("  Completed: %d\n", stats["completed"])
	fmt.Printf("  In Progress: %d\n", stats["in_progress"])
	fmt.Printf("  Pending: %d\n", stats["pending"])
	if stats["blocked"] > 0 {
		fmt.Printf("  Blocked: %d\n", stats["blocked"])
	}
}
