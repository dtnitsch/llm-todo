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
	root.AddCommand(deleteCmd())
	root.AddCommand(blockCmd())
	root.AddCommand(noteCmd())
	root.AddCommand(skipCmd())
	root.AddCommand(demoteCmd())
	root.AddCommand(promoteCmd())
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

			// Get upcoming tasks (next 2-3 after current)
			upcomingTasks, _ := mgr.GetUpcomingTasks(sessionID, 3)

			output := &todo.NextOutput{
				Task:           task,
				Session:        session,
				TotalTasks:     stats["total"],
				CompletedTasks: stats["completed"],
				UpcomingTasks:  upcomingTasks,
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

func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <task-ids>",
		Short: "Permanently delete task(s)",
		Long:  "Permanently delete tasks from the database. This cannot be undone.\n\nAliases: rm, remove",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			_ = getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Parse task IDs
			ids, err := todo.ParseTaskIDs(args[0])
			if err != nil {
				return err
			}

			// Delete tasks
			if err := mgr.BatchDeleteTasks(ids); err != nil {
				return err
			}

			if len(ids) == 1 {
				fmt.Printf("Deleted task #%d\n", ids[0])
			} else {
				fmt.Printf("Deleted %d tasks: %v\n", len(ids), ids)
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

func skipCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip [task-ids]",
		Short: "Skip task(s) (move to back of priority queue)",
		Long:  "Defer task(s) by adding 500 to priority_order, moving them to the back of their current priority level.\n\nNo args: skip current task\nWith args: skip specified task IDs (comma-separated)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			var ids []int

			// No args: skip current task
			if len(args) == 0 {
				task, err := mgr.GetNextTask(sessionID)
				if err != nil {
					return err
				}
				ids = []int{task.ID}
			} else {
				// Parse specified task IDs
				var err error
				ids, err = todo.ParseTaskIDs(args[0])
				if err != nil {
					return err
				}
			}

			// Add 500 to priority_order for each task
			for _, id := range ids {
				task, err := mgr.GetTask(id)
				if err != nil {
					return err
				}
				newOrder := task.PriorityOrder + 500
				if err := mgr.UpdateTask(id, map[string]interface{}{"priority_order": newOrder}); err != nil {
					return err
				}
			}

			if len(ids) == 1 {
				fmt.Printf("⏭️  Skipped task #%d\n", ids[0])
			} else {
				fmt.Printf("⏭️  Skipped %d tasks: %v\n", len(ids), ids)
			}

			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}

func demoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demote [task-ids]",
		Short: "Demote task(s) to lower priority (p0→p1, p1→p2, etc.)",
		Long:  "Reduce priority level by one (p0→p1, p1→p2, p2→p3, p3→p4). Priority order is preserved.\n\nNo args: demote current task\nWith args: demote specified task IDs (comma-separated)\n\nNote: Tasks at p4 cannot be demoted further.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			var ids []int

			// No args: demote current task
			if len(args) == 0 {
				task, err := mgr.GetNextTask(sessionID)
				if err != nil {
					return err
				}
				ids = []int{task.ID}
			} else {
				// Parse specified task IDs
				var err error
				ids, err = todo.ParseTaskIDs(args[0])
				if err != nil {
					return err
				}
			}

			// Priority mapping
			priorityMap := map[string]string{
				"p0": "p1",
				"p1": "p2",
				"p2": "p3",
				"p3": "p4",
			}

			// Demote each task
			for _, id := range ids {
				task, err := mgr.GetTask(id)
				if err != nil {
					return err
				}

				newPriority, ok := priorityMap[task.Priority]
				if !ok {
					return fmt.Errorf("task #%d is already at lowest priority (p4). Use 'llmtodo priority %d <order>' to adjust ordering", id, id)
				}

				if err := mgr.UpdateTask(id, map[string]interface{}{"priority": newPriority}); err != nil {
					return err
				}
			}

			if len(ids) == 1 {
				fmt.Printf("⬇️  Demoted task #%d\n", ids[0])
			} else {
				fmt.Printf("⬇️  Demoted %d tasks: %v\n", len(ids), ids)
			}

			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}

func promoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote [task-ids]",
		Short: "Promote task(s) to higher priority (p1→p0, p2→p1, etc.)",
		Long:  "Increase priority level by one (p1→p0, p2→p1, p3→p2, p4→p3). Priority order is preserved.\n\nNo args: promote current task\nWith args: promote specified task IDs (comma-separated)\n\nNote: Tasks at p0 cannot be promoted further.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionOverride, _ := cmd.Flags().GetString("session")
			sessionID := getSessionIDWithOverride(sessionOverride)
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			var ids []int

			// No args: promote current task
			if len(args) == 0 {
				task, err := mgr.GetNextTask(sessionID)
				if err != nil {
					return err
				}
				ids = []int{task.ID}
			} else {
				// Parse specified task IDs
				var err error
				ids, err = todo.ParseTaskIDs(args[0])
				if err != nil {
					return err
				}
			}

			// Priority mapping
			priorityMap := map[string]string{
				"p1": "p0",
				"p2": "p1",
				"p3": "p2",
				"p4": "p3",
			}

			// Promote each task
			for _, id := range ids {
				task, err := mgr.GetTask(id)
				if err != nil {
					return err
				}

				newPriority, ok := priorityMap[task.Priority]
				if !ok {
					return fmt.Errorf("task #%d is already at highest priority (p0). Use 'llmtodo priority %d <order>' to adjust ordering", id, id)
				}

				if err := mgr.UpdateTask(id, map[string]interface{}{"priority": newPriority}); err != nil {
					return err
				}
			}

			if len(ids) == 1 {
				fmt.Printf("⬆️  Promoted task #%d\n", ids[0])
			} else {
				fmt.Printf("⬆️  Promoted %d tasks: %v\n", len(ids), ids)
			}

			return nil
		},
	}

	cmd.Flags().String("session", "", "Query a different session")
	return cmd
}
