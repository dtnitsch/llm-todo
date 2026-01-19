package formatter

import (
	"fmt"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// PrintMinimal prints minimal task list (ID + status icon + title)
func PrintMinimal(tasks []*todo.Task, filter string) {
	if len(tasks) == 0 {
		fmt.Printf("No tasks found for: %s\n", filter)
		return
	}

	fmt.Printf("%s tasks (%d total):\n", filter, len(tasks))
	for _, task := range tasks {
		icon := statusIcon(task.Status)
		fmt.Printf("%s %d  %s\n", icon, task.ID, task.Task)
	}
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
