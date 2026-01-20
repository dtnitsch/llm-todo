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
	rootCmd.AddCommand(sessionCmd())
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

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session [show|goal]",
		Short: "View or update session context",
		Example: `  todo session                              # Show current session
  todo session show llm-todo                # Show specific session
  todo session goal "Refactor auth system"  # Set goal for current session
  todo session goal llm-todo "Build API"    # Set goal for specific session`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// No args: show current session
			if len(args) == 0 {
				return showCurrentSession()
			}

			// First arg is subcommand
			subcommand := args[0]

			switch subcommand {
			case "show":
				if len(args) < 2 {
					return showCurrentSession()
				}
				return showSession(args[1])

			case "goal":
				if len(args) < 2 {
					return fmt.Errorf("usage: todo session goal [session-id] \"<goal>\"")
				}

				// Check if second arg is session ID or goal
				sessionID := getSessionID()
				goalText := args[1]

				// If 3 args, first is session ID, second is goal
				if len(args) == 3 {
					sessionID = args[1]
					goalText = args[2]
				}

				return setSessionGoal(sessionID, goalText)

			default:
				return fmt.Errorf("unknown subcommand: %s", subcommand)
			}
		},
	}

	return cmd
}

func showCurrentSession() error {
	sessionID := getSessionID()
	return showSession(sessionID)
}

func showSession(sessionID string) error {
	mgr, err := todo.NewManager("")
	if err != nil {
		return err
	}
	defer mgr.Close()

	session, err := mgr.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	stats, _ := mgr.GetStats(sessionID)

	fmt.Printf("Session: %s (%s)\n", session.ID, session.Type)
	if session.Goal != "" {
		fmt.Printf("Goal: %s\n", session.Goal)
	} else {
		fmt.Printf("Goal: (not set)\n")
	}

	if session.SuccessCriteria != "" {
		fmt.Printf("Success criteria: %s\n", session.SuccessCriteria)
	}

	fmt.Printf("\nProgress:\n")
	fmt.Printf("  Total: %d\n", stats["total"])
	fmt.Printf("  Completed: %d\n", stats["completed"])
	fmt.Printf("  In Progress: %d\n", stats["in_progress"])
	fmt.Printf("  Pending: %d\n", stats["pending"])
	if stats["blocked"] > 0 {
		fmt.Printf("  Blocked: %d\n", stats["blocked"])
	}

	return nil
}

func setSessionGoal(sessionID, goal string) error {
	mgr, err := todo.NewManager("")
	if err != nil {
		return err
	}
	defer mgr.Close()

	// Verify session exists
	_, err = mgr.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Update goal
	updates := map[string]interface{}{"goal": goal}
	if err := mgr.UpdateSession(sessionID, updates); err != nil {
		return err
	}

	fmt.Printf("✓ Updated goal for session: %s\n", sessionID)
	fmt.Printf("  Goal: %s\n", goal)
	return nil
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
