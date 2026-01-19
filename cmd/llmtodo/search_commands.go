package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(searchCmd())
}

func searchCmd() *cobra.Command {
	var onlyPending bool

	cmd := &cobra.Command{
		Use:   "search <keyword>",
		Short: "Search tasks (includes completed by default)",
		Example: `  todo search "auth"
  todo search "database schema"
  todo search "bug" --pending  # Only pending tasks`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := getSessionID()
			query := args[0]

			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			includeCompleted := !onlyPending
			matches, err := mgr.SearchTasks(sessionID, query, includeCompleted)
			if err != nil {
				return err
			}

			if len(matches) == 0 {
				fmt.Printf("No tasks found for: %s\n", query)
				return nil
			}

			fmt.Printf("Found %d tasks matching '%s':\n\n", len(matches), query)

			for _, task := range matches {
				// Show status icon
				icon := getStatusIcon(task.Status)

				// Show task
				fmt.Printf("%s #%d: %s\n", icon, task.ID, task.Task)
				fmt.Printf("   Status: %s, Priority: %s\n", task.Status, task.Priority)

				// Show notes snippet if present
				if task.Notes != "" {
					snippet := truncate(task.Notes, 100)
					fmt.Printf("   Notes: %s\n", snippet)
				}

				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&onlyPending, "pending", false, "Search only pending tasks (excludes completed)")

	return cmd
}

func getStatusIcon(status string) string {
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
