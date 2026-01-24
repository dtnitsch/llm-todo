package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/exporter"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(exportCmd())
}

func exportCmd() *cobra.Command {
	var sessionID, dir string
	var scaffold, includeDone bool

	cmd := &cobra.Command{
		Use:   "export <file.yaml>",
		Short: "Export tasks to YAML file(s)",
		Long: `Export tasks to YAML file(s).

Usage:
  llmtodo export <file.yaml>        # Export current session
  llmtodo export --dir todo/        # Split by priority
  llmtodo export --scaffold todo/   # Create empty structure

Examples:
  llmtodo export tasks.yaml
  llmtodo export --dir todo/
  llmtodo export --scaffold todo/`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get session ID
			sessionOverride, _ := cmd.Flags().GetString("session")
			if sessionOverride != "" {
				sessionID = sessionOverride
			} else {
				sessionID = getSessionID()
			}

			// Handle scaffold mode
			if scaffold {
				if len(args) == 0 && dir == "" {
					return fmt.Errorf("provide directory path for scaffold")
				}
				scaffoldDir := dir
				if scaffoldDir == "" {
					scaffoldDir = args[0]
				}
				return exporter.GenerateScaffold(scaffoldDir)
			}

			// Handle export
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Export by priority to directory
			if dir != "" {
				return exporter.ExportByPriority(mgr, sessionID, dir, includeDone)
			}

			// Export to single file
			if len(args) == 0 {
				return fmt.Errorf("provide file path or --dir")
			}
			return exporter.ExportToYAML(mgr, sessionID, args[0])
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Session to export (default: current)")
	cmd.Flags().StringVar(&dir, "dir", "", "Export to directory (split by priority)")
	cmd.Flags().BoolVar(&scaffold, "scaffold", false, "Create empty structure with examples")
	cmd.Flags().BoolVar(&includeDone, "include-done", false, "Include completed tasks in export")

	return cmd
}
