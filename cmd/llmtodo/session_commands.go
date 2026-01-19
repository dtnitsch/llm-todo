package main

import (
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(sessionsCmd())
	rootCmd.AddCommand(useCmd())
}

func sessionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sessions",
		Short: "List all sessions with progress",
		Example: `  todo sessions
  todo sessions | grep llm-todo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			sessions, err := mgr.ListSessions()
			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found")
				return nil
			}

			// Show current session
			currentSession, _ := todo.GetCurrentSession()
			if currentSession != "" {
				fmt.Printf("Active session: %s\n\n", currentSession)
			}

			// Table output
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintln(w, "Session\tType\tTotal\tPending\tIn Progress\tCompleted\tBlocked\tUpdated")
			fmt.Fprintln(w, "---\t---\t---\t---\t---\t---\t---\t---")

			for _, s := range sessions {
				active := ""
				if s.ID == currentSession {
					active = "*"
				}

				goal := s.Goal
				if len(goal) > 30 {
					goal = goal[:27] + "..."
				}

				fmt.Fprintf(w, "%s%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n",
					active, s.ID, s.Type, s.Total, s.Pending, s.InProgress, s.Completed, s.Blocked, s.UpdatedAt)
			}

			return nil
		},
	}
}

func useCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <session-id>",
		Short: "Switch to a different session",
		Example: `  todo use llm-todo
  todo use my-project`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			// Verify session exists
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			_, err = mgr.GetSession(sessionID)
			if err != nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			// Set as current
			if err := todo.SetCurrentSession(sessionID); err != nil {
				return err
			}

			fmt.Printf("✓ Switched to session: %s\n", sessionID)
			fmt.Printf("\nNext: todo next\n")
			return nil
		},
	}
}
