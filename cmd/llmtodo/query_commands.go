package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/formatter"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addQueryCommands(root *cobra.Command) {
	root.AddCommand(getCmd())
	root.AddCommand(showCmd())
	root.AddCommand(statusCmd())
}

func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <priority|status>",
		Short: "List tasks (minimal: ID + title)",
		Long: `List tasks by priority or status.

PRIORITY FILTERS (excludes completed by default):
  p0, p1, p2, p3, p4

STATUS FILTERS:
  pending, in_progress, completed, blocked

FLAGS:
  --all      Include completed tasks (for priority filters)
  --full     Show all tasks (no 10-item limit on queued)
  --session  Query a different session (default: current session)

By default, priority filters (p0, p1, etc.) exclude completed tasks.
Use --all to include them. Status filters always show all matching tasks.`,
		Example: `  llmtodo get p0                  # High-priority pending tasks
  llmtodo get pending             # All pending tasks
  llmtodo get completed           # All completed tasks
  llmtodo get blocked             # All blocked tasks
  llmtodo get p0 --all            # p0 tasks including completed
  llmtodo get p0 --full           # All p0 pending (no 10-item limit)
  llmtodo get p0 --all --full     # All p0 tasks, full list
  llmtodo get p0 --session other  # p0 tasks from 'other' session`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			arg := args[0]
			filters := make(map[string]string)

			// Check if it's priority or status
			isPriority := false
			if arg == "p0" || arg == "p1" || arg == "p2" || arg == "p3" || arg == "p4" {
				filters["priority"] = arg
				isPriority = true
			} else {
				filters["status"] = arg
			}

			// Get flags
			showAll, _ := cmd.Flags().GetBool("all")
			showFull, _ := cmd.Flags().GetBool("full")

			// For priority filters, exclude completed by default unless --all
			if isPriority && !showAll {
				// We'll filter out completed in the formatter instead
				// to get accurate counts
			}

			tasks, err := mgr.ListTasks(sessionID, filters)
			if err != nil {
				return err
			}

			formatter.PrintMinimal(tasks, arg, showAll, showFull, isPriority)
			return nil
		},
	}

	cmd.Flags().Bool("all", false, "Include completed tasks")
	cmd.Flags().Bool("full", false, "Show full lists (no 10-item limit)")
	cmd.Flags().String("session", "", "Query a different session")

	return cmd
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show full task details",
		Long: `Display complete details for a specific task.

Shows all available information including:
- Task title and status
- Priority and effort estimate
- Instructions (must_do, must_not_do)
- Related files
- Notes and blocking reasons`,
		Example: `  llmtodo show 144               # Show task #144
  llmtodo fetch 144              # Alias for show
  llmtodo read 144               # Alias for show`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			var taskID int
			fmt.Sscanf(args[0], "%d", &taskID)

			task, err := mgr.GetTask(taskID)
			if err != nil {
				return err
			}

			formatter.PrintFull(task)
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show session progress summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			session, err := mgr.GetSession(sessionID)
			if err != nil {
				return err
			}

			stats, err := mgr.GetStats(sessionID)
			if err != nil {
				return err
			}

			formatter.PrintStatus(session, stats)
			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")

	return cmd
}
