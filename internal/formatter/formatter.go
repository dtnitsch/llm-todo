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

	// Header
	completed := output.CompletedTasks
	total := output.TotalTasks
	b.WriteString(fmt.Sprintf("🎯 NEXT: %s (task %d/%d)\n\n", task.Task, completed+1, total))

	// SUGGESTIONS (if any)
	if len(suggestions) > 0 {
		b.WriteString("💡 SUGGESTIONS:\n")
		for _, sug := range suggestions {
			b.WriteString(fmt.Sprintf("  • %s\n", sug.Message))
		}
		b.WriteString("\n")
	}

	// INSTRUCTIONS (must_do / must_not_do)
	if task.Instructions != "" {
		var instructions map[string][]string
		if err := json.Unmarshal([]byte(task.Instructions), &instructions); err == nil {
			if mustDo := instructions["must_do"]; len(mustDo) > 0 {
				b.WriteString("INSTRUCTIONS:\n")
				for _, item := range mustDo {
					b.WriteString(fmt.Sprintf("  ✓ %s\n", item))
				}
				b.WriteString("\n")
			}
			if mustNotDo := instructions["must_not_do"]; len(mustNotDo) > 0 {
				b.WriteString("MUST NOT:\n")
				for _, item := range mustNotDo {
					b.WriteString(fmt.Sprintf("  ✗ %s\n", item))
				}
				b.WriteString("\n")
			}
		}
	}

	// FILES
	if task.Files != "" && task.Files != "[]" {
		var files []string
		if err := json.Unmarshal([]byte(task.Files), &files); err == nil && len(files) > 0 {
			b.WriteString(fmt.Sprintf("FILES: %s\n\n", strings.Join(files, ", ")))
		}
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
		b.WriteString(fmt.Sprintf("⏳ WAITING ON: %s\n\n", task.WaitingOn))
	}

	// OUTPUT/AUDIENCE (deliverable mode)
	if task.Output != "" {
		b.WriteString(fmt.Sprintf("📤 OUTPUT: %s\n", task.Output))
		if task.Audience != "" {
			b.WriteString(fmt.Sprintf("👥 AUDIENCE: %s\n", task.Audience))
		}
		b.WriteString("\n")
	}

	// NOTES
	if task.Notes != "" {
		b.WriteString(fmt.Sprintf("💡 NOTES:\n%s\n\n", task.Notes))
	}

	// Footer
	b.WriteString("After completing: todo done\n")

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

	if task.BlockingReason != "" {
		fmt.Printf("\n⚠️  Blocked: %s\n", task.BlockingReason)
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
