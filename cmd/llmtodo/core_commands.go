package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/formatter"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addCoreCommands(root *cobra.Command) {
	root.AddCommand(nextCmd())
	root.AddCommand(doneCmd())
	root.AddCommand(blockCmd())
	root.AddCommand(noteCmd())
}

func nextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show next task with full details",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			task, err := mgr.GetNextTask(sessionID)
			if err != nil {
				return err
			}

			session, _ := mgr.GetSession(sessionID)
			stats, _ := mgr.GetStats(sessionID)

			output := &todo.NextOutput{
				Task:           task,
				Session:        session,
				TotalTasks:     stats["total"],
				CompletedTasks: stats["completed"],
			}

			// Get suggestions
			suggestions, _ := mgr.GetSuggestions(sessionID)

			fmt.Print(formatter.FormatNext(output, suggestions))
			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}

func doneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [task-ids]",
		Short: "Mark task(s) as completed",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// No args: complete current task
			if len(args) == 0 {
				task, err := mgr.GetNextTask(sessionID)
				if err != nil {
					return err
				}
				if err := mgr.UpdateTaskStatus(task.ID, "completed"); err != nil {
					return err
				}
				fmt.Printf("✅ Completed task #%d\n", task.ID)

				// Auto-track modified files
				trackedFiles, _ := mgr.TrackModifiedFiles(task.ID)
				if len(trackedFiles) > 0 {
					fmt.Printf("📁 Auto-tracked files: %s\n", strings.Join(trackedFiles, ", "))
				}

				return checkUnblockedTasks(mgr, sessionID, task.ID)
			}

			// Parse and complete multiple
			ids, err := todo.ParseTaskIDs(args[0])
			if err != nil {
				return err
			}

			if err := mgr.BatchUpdateStatus(ids, "completed"); err != nil {
				return err
			}

			fmt.Printf("✅ Completed %d tasks: %v\n", len(ids), ids)

			// Auto-track files for first completed task
			if len(ids) > 0 {
				trackedFiles, _ := mgr.TrackModifiedFiles(ids[0])
				if len(trackedFiles) > 0 {
					fmt.Printf("📁 Auto-tracked files: %s\n", strings.Join(trackedFiles, ", "))
				}
			}

			// Check for unblocked tasks
			for _, id := range ids {
				checkUnblockedTasks(mgr, sessionID, id)
			}

			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}

func blockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block <task-ids> <reason>",
		Short: "Mark task(s) as blocked",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			ids, err := todo.ParseTaskIDs(args[0])
			if err != nil {
				return err
			}

			reason := args[1]

			for _, id := range ids {
				updates := map[string]interface{}{
					"status":          "blocked",
					"blocking_reason": reason,
				}
				if err := mgr.UpdateTask(id, updates); err != nil {
					return err
				}
			}

			fmt.Printf("⚠️  Blocked %d tasks: %v\n", len(ids), ids)
			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}

func noteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note <task-ids> <note>",
		Short: "Add note to task(s)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			ids, err := todo.ParseTaskIDs(args[0])
			if err != nil {
				return err
			}

			note := args[1]

			if err := mgr.BatchAddNote(ids, note); err != nil {
				return err
			}

			fmt.Printf("💡 Added note to %d tasks\n", len(ids))
			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}
