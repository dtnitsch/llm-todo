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

			fmt.Fprintln(w, "Session\tType\tGoal\tTotal\tPending\tIn Progress\tCompleted\tBlocked\tUpdated")
			fmt.Fprintln(w, "---\t---\t---\t---\t---\t---\t---\t---\t---")

			for _, s := range sessions {
				active := ""
				if s.ID == currentSession {
					active = "*"
				}

				goal := s.Goal
				if len(goal) > 30 {
					goal = goal[:27] + "..."
				}

				fmt.Fprintf(w, "%s%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n",
					active, s.ID, s.Type, goal, s.Total, s.Pending, s.InProgress, s.Completed, s.Blocked, s.UpdatedAt)
			}

			return nil
		},
	}
}

func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session [show|goal]",
		Short: "View or update session context",
		Long: `View or update session context including goal, progress, and recent tasks.

Shows session overview with:
- Session goal and success criteria
- Progress statistics (completed/pending/blocked)
- Recently completed tasks (last 5)
- Next pending tasks (next 5 by priority)
- All blocked tasks

Use --all flag to see all tasks instead of recent summaries.`,
		Example: `  llmtodo session                           # Show current session with recent tasks
  llmtodo session --all                     # Show current session with all tasks
  llmtodo session show llm-todo             # Show specific session
  llmtodo session show llm-todo --all       # Show specific session with all tasks
  llmtodo session goal "Refactor auth"      # Set goal for current session
  llmtodo session goal llm-todo "Build API" # Set goal for specific session`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			showAll, _ := cmd.Flags().GetBool("all")

			// No args: show current session
			if len(args) == 0 {
				return showCurrentSession(showAll)
			}

			// First arg is subcommand
			subcommand := args[0]

			switch subcommand {
			case "show":
				if len(args) < 2 {
					return showCurrentSession(showAll)
				}
				return showSession(args[1], showAll)

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

	cmd.Flags().Bool("all", false, "Show all tasks instead of recent summaries")

	return cmd
}

func showCurrentSession(showAll bool) error {
	sessionID := getSessionID()
	return showSession(sessionID, showAll)
}

func showSession(sessionID string, showAll bool) error {
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

	// Header
	fmt.Printf("Session: %s (%s)\n", session.ID, session.Type)
	if session.Goal != "" {
		fmt.Printf("Goal: %s\n", session.Goal)
	} else {
		fmt.Printf("Goal: (not set)\n")
	}

	if session.SuccessCriteria != "" {
		fmt.Printf("Success criteria: %s\n", session.SuccessCriteria)
	}

	// Progress
	total := stats["total"]
	completed := stats["completed"]
	percentage := 0
	if total > 0 {
		percentage = (completed * 100) / total
	}
	fmt.Printf("\nProgress: %d/%d completed (%d%%)\n", completed, total, percentage)
	if stats["in_progress"] > 0 {
		fmt.Printf("  In Progress: %d\n", stats["in_progress"])
	}
	if stats["pending"] > 0 {
		fmt.Printf("  Pending: %d\n", stats["pending"])
	}
	if stats["blocked"] > 0 {
		fmt.Printf("  Blocked: %d\n", stats["blocked"])
	}

	// Show tasks
	if showAll {
		printAllSessionTasks(mgr, sessionID)
	} else {
		printRecentSessionTasks(mgr, sessionID)
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

func printRecentSessionTasks(mgr *todo.Manager, sessionID string) {
	// Recently completed (last 5)
	completedTasks, _ := mgr.ListTasks(sessionID, map[string]string{"status": "completed"})
	if len(completedTasks) > 0 {
		fmt.Printf("\nRECENTLY COMPLETED:\n")
		start := len(completedTasks) - 5
		if start < 0 {
			start = 0
		}
		for i := len(completedTasks) - 1; i >= start; i-- {
			t := completedTasks[i]
			fmt.Printf("  [done] %d %s\n", t.ID, t.Task)
		}
	}

	// Next pending (first 5 by priority)
	pendingTasks, _ := mgr.ListTasks(sessionID, map[string]string{"status": "pending"})
	if len(pendingTasks) > 0 {
		fmt.Printf("\nNEXT UP (p0):\n")
		count := 0
		for _, t := range pendingTasks {
			if t.Priority == "p0" && count < 5 {
				fmt.Printf("  %d  %s\n", t.ID, t.Task)
				count++
			}
		}
		if count == 0 {
			// Show first 5 pending regardless of priority
			limit := 5
			if len(pendingTasks) < limit {
				limit = len(pendingTasks)
			}
			for i := 0; i < limit; i++ {
				t := pendingTasks[i]
				fmt.Printf("  %d  %s\n", t.ID, t.Task)
			}
		}
	}

	// All blocked
	blockedTasks, _ := mgr.ListTasks(sessionID, map[string]string{"status": "blocked"})
	if len(blockedTasks) > 0 {
		fmt.Printf("\nBLOCKED:\n")
		for _, t := range blockedTasks {
			fmt.Printf("  %d  %s\n", t.ID, t.Task)
			if t.BlockingReason != "" {
				fmt.Printf("      Reason: %s\n", t.BlockingReason)
			}
		}
	}
}

func printAllSessionTasks(mgr *todo.Manager, sessionID string) {
	// All tasks grouped by status
	allTasks, _ := mgr.ListTasks(sessionID, map[string]string{})

	statuses := []string{"pending", "in_progress", "blocked", "completed"}
	statusLabels := map[string]string{
		"pending":     "PENDING",
		"in_progress": "IN PROGRESS",
		"blocked":     "BLOCKED",
		"completed":   "COMPLETED",
	}

	for _, status := range statuses {
		var tasks []*todo.Task
		for _, t := range allTasks {
			if t.Status == status {
				tasks = append(tasks, t)
			}
		}

		if len(tasks) > 0 {
			fmt.Printf("\n%s:\n", statusLabels[status])
			for _, t := range tasks {
				prefix := ""
				if status == "completed" {
					prefix = "[done] "
				} else if status == "blocked" {
					prefix = "[blocked] "
				}
				fmt.Printf("  %s%d %s", prefix, t.ID, t.Task)
				if t.Priority != "" {
					fmt.Printf(" [%s]", t.Priority)
				}
				fmt.Println()
				if status == "blocked" && t.BlockingReason != "" {
					fmt.Printf("      Reason: %s\n", t.BlockingReason)
				}
			}
		}
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
