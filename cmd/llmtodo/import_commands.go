package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/importer"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(importCmd())
}

func importCmd() *cobra.Command {
	var sessionID, dir string

	cmd := &cobra.Command{
		Use:   "import <file.yaml>",
		Short: "Import tasks from YAML file",
		Example: `  todo import tasks.yaml
  todo import --dir todo/  # Import all p0-p4 files
  todo import --session mysession tasks.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Use current directory name as session ID if not provided
			if sessionID == "" {
				sessionID = filepath.Base(mustGetwd())
			}

			// Ensure session exists
			session, err := mgr.GetOrCreateSession(sessionID, "code")
			if err != nil {
				return err
			}

			var count int
			var goal string

			// Import from directory
			if dir != "" {
				count, goal, err = importer.ImportFromDirectory(mgr, sessionID, dir)
			} else if len(args) > 0 {
				// Import single file
				count, goal, err = importer.ImportFromYAML(mgr, sessionID, args[0])
			} else {
				return fmt.Errorf("provide file path or --dir")
			}

			if err != nil {
				return err
			}

			// Update session goal if found in YAML
			if goal != "" {
				updates := map[string]interface{}{"goal": goal}
				if err := mgr.UpdateSession(sessionID, updates); err != nil {
					return err
				}
			}

			// Auto-switch to the imported session
			if err := todo.SetCurrentSession(sessionID); err != nil {
				return err
			}

			fmt.Printf("✓ Imported %d tasks into session: %s\n", count, sessionID)
			fmt.Printf("✓ Switched to session: %s\n", sessionID)
			if goal != "" {
				fmt.Printf("\nGoal: %s\n", goal)
			}

			// Show priority breakdown
			priorityStats, _ := mgr.GetPriorityStats(sessionID)
			if len(priorityStats) > 0 {
				fmt.Printf("\nBreakdown by priority:\n")
				for _, p := range []string{"p0", "p1", "p2", "p3", "p4"} {
					if count, ok := priorityStats[p]; ok && count > 0 {
						label := "high priority"
						if p == "p1" {
							label = "important"
						} else if p == "p2" {
							label = "medium"
						} else if p == "p3" {
							label = "low"
						} else if p == "p4" {
							label = "optional"
						}
						fmt.Printf("  %s: %d tasks (%s)\n", p, count, label)
					}
				}
			}

			// Show actionable next steps
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  llmtodo next           # See next task\n")
			if priorityStats["p0"] > 0 {
				fmt.Printf("  llmtodo get p0         # See all p0 tasks\n")
			}
			if goal == "" && session.Goal == "" {
				fmt.Printf("  llmtodo session goal %s \"<description>\"  # Add session goal\n", sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID (defaults to current directory)")
	cmd.Flags().StringVar(&dir, "dir", "", "Import all p0-p4.yaml files from directory")

	return cmd
}
