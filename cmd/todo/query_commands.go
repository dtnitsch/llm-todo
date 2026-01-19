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
	return &cobra.Command{
		Use:   "get <priority|status>",
		Short: "List tasks (minimal: ID + title)",
		Example: `  todo get p0
  todo get pending
  todo get completed`,
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
			if arg == "p0" || arg == "p1" || arg == "p2" || arg == "p3" || arg == "p4" {
				filters["priority"] = arg
			} else {
				filters["status"] = arg
			}

			tasks, err := mgr.ListTasks(sessionID, filters)
			if err != nil {
				return err
			}

			formatter.PrintMinimal(tasks, arg)
			return nil
		},
	}
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
