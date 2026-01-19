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
		Example: `  todo get p0
  todo get pending
  todo get completed
  todo get p0 --all        # include completed
  todo get p0 --full       # show all queued (no 10-item limit)
  todo get p0 --all --full # show everything`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := getSessionID()
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

	return cmd
}

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <task-id>",
		Short: "Show full task details",
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
			sessionID := getSessionID()
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

	// Default run when no args (alias for status)
	rootCmd.RunE = cmd.RunE

	return cmd
}
