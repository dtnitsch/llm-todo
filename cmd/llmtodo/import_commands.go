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

			fmt.Printf("✓ Imported %d tasks into session: %s\n", count, sessionID)
			if goal != "" {
				fmt.Printf("  Goal: %s\n", goal)
			}

			stats, _ := mgr.GetStats(sessionID)
			fmt.Printf("\nSession stats:\n")
			fmt.Printf("  Total: %d\n", stats["total"])
			fmt.Printf("  Pending: %d\n", stats["pending"])

			// Show suggestion if goal not in YAML and not already in session
			if goal == "" && session.Goal == "" {
				printGoalSuggestion(sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID (defaults to current directory)")
	cmd.Flags().StringVar(&dir, "dir", "", "Import all p0-p4.yaml files from directory")

	return cmd
}
