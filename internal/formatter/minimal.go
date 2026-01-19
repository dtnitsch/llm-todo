package formatter

import (
	"fmt"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// PrintMinimal prints minimal task list (ID + status icon + title)
// For priority filters with showAll=false, groups by ACTIVE/QUEUED and excludes completed
// Limits QUEUED to 10 items unless showFull=true
func PrintMinimal(tasks []*todo.Task, filter string, showAll bool, showFull bool, isPriority bool) {
	if len(tasks) == 0 {
		fmt.Printf("No tasks found for: %s\n", filter)
		return
	}

	// If not a priority filter, use old behavior
	if !isPriority {
		fmt.Printf("%s tasks (%d total):\n", filter, len(tasks))
		for _, task := range tasks {
			icon := statusIcon(task.Status)
			fmt.Printf("%s %d  %s\n", icon, task.ID, task.Task)
		}
		return
	}

	// Group tasks by status
	var active, queued, completed, blocked []*todo.Task
	for _, task := range tasks {
		switch task.Status {
		case "in_progress":
			active = append(active, task)
		case "pending":
			queued = append(queued, task)
		case "completed":
			if showAll {
				completed = append(completed, task)
			}
		case "blocked":
			blocked = append(blocked, task)
		}
	}

	// Build header with counts
	header := fmt.Sprintf("%s:", filter)
	parts := []string{}
	if len(active) > 0 {
		parts = append(parts, fmt.Sprintf("%d active", len(active)))
	}
	if len(queued) > 0 {
		parts = append(parts, fmt.Sprintf("%d queued", len(queued)))
	}
	if len(blocked) > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked", len(blocked)))
	}
	if len(completed) > 0 {
		parts = append(parts, fmt.Sprintf("%d completed", len(completed)))
	}

	if len(parts) > 0 {
		header += " " + joinParts(parts)
	}

	fmt.Printf("%s\n\n", header)

	// Print ACTIVE section
	if len(active) > 0 {
		fmt.Println("ACTIVE:")
		for _, task := range active {
			fmt.Printf("  %d  %s\n", task.ID, task.Task)
		}
		fmt.Println()
	}

	// Print QUEUED section (with limiting)
	if len(queued) > 0 {
		queuedLimit := len(queued)
		limitApplied := false
		if !showFull && len(queued) > 10 {
			queuedLimit = 10
			limitApplied = true
		}

		if limitApplied {
			fmt.Printf("QUEUED (showing %d/%d, use --full for all):\n", queuedLimit, len(queued))
		} else {
			fmt.Println("QUEUED:")
		}

		for i := 0; i < queuedLimit; i++ {
			fmt.Printf("  %d  %s\n", queued[i].ID, queued[i].Task)
		}
		fmt.Println()
	}

	// Print BLOCKED section
	if len(blocked) > 0 {
		fmt.Println("BLOCKED:")
		for _, task := range blocked {
			fmt.Printf("  %d  %s\n", task.ID, task.Task)
		}
		fmt.Println()
	}

	// Print COMPLETED section (with limiting if --all)
	if len(completed) > 0 {
		completedLimit := len(completed)
		limitApplied := false
		if !showFull && len(completed) > 10 {
			completedLimit = 10
			limitApplied = true
		}

		if limitApplied {
			fmt.Printf("COMPLETED (showing %d/%d, use --full for all):\n", completedLimit, len(completed))
		} else {
			fmt.Println("COMPLETED:")
		}

		for i := 0; i < completedLimit; i++ {
			fmt.Printf("  %d  %s\n", completed[i].ID, completed[i].Task)
		}
	}
}

func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ", "
		}
		result += part
	}
	return result
}

func statusIcon(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "in_progress":
		return "→"
	case "blocked":
		return "⚠"
	case "pending":
		return "-"
	default:
		return " "
	}
}
