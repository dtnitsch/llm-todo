package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/importer"
	"github.com/dtnitsch/llm-todo/internal/templates"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(importCmd())
}

func importCmd() *cobra.Command {
	var sessionID, dir string
	var showTemplate, showTemplateFull, validate bool

	cmd := &cobra.Command{
		Use:   "import <file.yaml>",
		Short: "Import tasks from YAML file",
		Long: `Import tasks from YAML file.

Usage:
  llmtodo import <file.yaml>
  llmtodo import --dir todo/        # Import p0-p4 files from directory
  llmtodo import --template         # Show minimal YAML template
  llmtodo import --template-full    # Show complete format
  llmtodo import --validate <file>  # Check if file is valid (no import)

Examples:
  llmtodo import tasks.yaml
  llmtodo import --dir todo/
  llmtodo import --template > tasks.yaml
  llmtodo import --validate tasks.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle template flags
			if showTemplate {
				fmt.Print(templates.MinimalImportTemplate())
				return nil
			}
			if showTemplateFull {
				fmt.Print(templates.FullImportTemplate())
				return nil
			}

			// Handle validate flag
			if validate {
				if len(args) == 0 {
					return fmt.Errorf("provide file path to validate")
				}
				return validateImportFile(args[0])
			}
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Determine session ID if not provided
			if sessionID == "" {
				// First try to get current session
				currentSession, err := todo.GetCurrentSession()
				if err == nil && currentSession != "" {
					sessionID = currentSession
				} else {
					// Fall back to directory name
					sessionID = filepath.Base(mustGetwd())
				}
			}

			// Ensure session exists
			session, err := mgr.GetOrCreateSession(sessionID, "code")
			if err != nil {
				return err
			}

			var count, createdCount, updatedCount int
			var goal string
			var mode string

			// Import from directory or file
			if dir != "" {
				count, goal, err = importer.ImportFromDirectory(mgr, sessionID, dir)
				mode = "create"
			} else if len(args) > 0 {
				// Detect import mode
				mode = detectImportMode(args[0])

				if mode == "hybrid" {
					// Hybrid mode: update tasks with IDs, create tasks without IDs
					createdCount, updatedCount, goal, err = importer.ImportAndUpdateFromYAML(mgr, sessionID, args[0])
					count = createdCount + updatedCount
				} else if mode == "update" {
					// Update existing tasks only
					count, goal, err = importer.UpdateTasksFromYAML(mgr, sessionID, args[0])
					updatedCount = count
				} else {
					// Create new tasks only
					count, goal, err = importer.ImportFromYAML(mgr, sessionID, args[0])
					createdCount = count
				}
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

			// Show different message based on mode
			if mode == "hybrid" {
				if createdCount > 0 && updatedCount > 0 {
					fmt.Printf("Created %d new tasks and updated %d existing tasks\n", createdCount, updatedCount)
				} else if createdCount > 0 {
					fmt.Printf("Created %d new tasks\n", createdCount)
				} else if updatedCount > 0 {
					fmt.Printf("Updated %d tasks from enrichment file\n", updatedCount)
				}
			} else if mode == "update" {
				fmt.Printf("Updated %d tasks from enrichment file\n", count)
			} else {
				fmt.Printf("Imported %d tasks into session: %s\n", count, sessionID)
			}
			fmt.Printf("Switched to session: %s\n", sessionID)
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
	cmd.Flags().BoolVar(&showTemplate, "template", false, "Output minimal YAML template")
	cmd.Flags().BoolVar(&showTemplateFull, "template-full", false, "Output complete YAML template with all fields")
	cmd.Flags().BoolVar(&validate, "validate", false, "Validate YAML file without importing")

	return cmd
}

// validateImportFile validates a YAML file without importing (parse-only, no DB access)
func validateImportFile(filePath string) error {
	// Detect import mode
	mode := detectImportMode(filePath)

	fmt.Printf("Validating: %s\n", filePath)
	if mode == "hybrid" {
		fmt.Println("   Mode: Hybrid (update tasks with IDs, create tasks without IDs)")
	} else if mode == "update" {
		fmt.Println("   Mode: Update (enrichment file with task IDs)")
	} else {
		fmt.Println("   Mode: Create (new tasks)")
	}
	fmt.Println()

	// Validate by parsing and checking fields (no DB access)
	count, goal, err := importer.ValidateYAMLFile(filePath)
	if err != nil {
		fmt.Println("ERROR: Validation FAILED\n")
		return err
	}

	// Success
	fmt.Println("Validation PASSED")
	fmt.Printf("   Would process: %d tasks\n", count)
	if goal != "" {
		fmt.Printf("   Goal: %s\n", goal)
	}
	fmt.Println()
	fmt.Println("HINT: File is valid. Run without --validate to import:")
	fmt.Printf("   llmtodo import %s\n", filePath)

	return nil
}

// detectImportMode checks the YAML file and returns the import mode
// Returns "create", "update", or "hybrid"
func detectImportMode(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "create"
	}

	content := string(data)

	// No IDs at all = create mode
	if !strings.Contains(content, "id: task-") {
		return "create"
	}

	// If file has "depends_on:" field, it's create mode (dependencies use logical IDs)
	// Enrichment files don't have depends_on
	if strings.Contains(content, "depends_on:") {
		return "create"
	}

	// Has some IDs - now check if SOME tasks don't have IDs (hybrid mode)
	// Simple approach: count lines with "id: task-" vs lines with "title:"
	lines := strings.Split(content, "\n")
	idCount := 0
	titleCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check for ID lines (can be "- id: task-1" or "id: task-1")
		if strings.Contains(trimmed, "id: task-") {
			idCount++
		}
		// Check for title lines (can be "- title:" or "title:")
		if strings.Contains(trimmed, "title:") && !strings.Contains(trimmed, "title: \"\"") {
			titleCount++
		}
	}

	// If we have more titles than IDs, some tasks don't have IDs (hybrid mode)
	hasTasksWithIDs := idCount > 0
	hasTasksWithoutIDs := titleCount > idCount

	// Determine mode
	if hasTasksWithIDs && hasTasksWithoutIDs {
		return "hybrid"
	} else if hasTasksWithIDs {
		return "update"
	} else {
		return "create"
	}
}
