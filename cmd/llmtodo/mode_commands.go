package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/exporter"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addModeCommands(root *cobra.Command) {
	root.AddCommand(quickCmd())
	root.AddCommand(codeCmd())
	root.AddCommand(researchCmd())
}

func quickCmd() *cobra.Command {
	var goal, name string

	cmd := &cobra.Command{
		Use:   "quick <task1> <task2> ...",
		Short: "Create quick session (3-5 tasks)",
		Long: `Create quick session (3-5 tasks).

Session naming:
  --name: Creates {directory}-{name} (e.g., llm-todo-auth-refactor)
  Auto:   Creates {directory}-{timestamp} (e.g., llm-todo-20260122-1430)

Tip: Use --name for focused work, skip for quick capture.`,
		Example: `  llmtodo quick "Fix bug" "Test" "PR" --name bug-1234
  llmtodo quick "Research enrichment"  # Auto: llm-todo-20260122-1430`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := generateSessionID(name)
			isAutoNamed := name == ""

			err := createSessionWithMetadata(sessionID, "quick", args, goal, "", "", "")
			if err != nil {
				return err
			}

			// Set as current session
			if err := todo.SetCurrentSession(sessionID); err != nil {
				return err
			}

			// Show naming hint if auto-generated
			if isAutoNamed {
				printNamingHint()
			}

			// Show suggestion if goal not provided
			if goal == "" {
				printGoalSuggestion(sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Session context (why these tasks exist)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name (creates {dir}-{name}, otherwise auto-timestamp)")
	return cmd
}

func codeCmd() *cobra.Command{
	var goal, boundaries, successCriteria, name string
	var skipPrompt bool

	cmd := &cobra.Command{
		Use:   "code <task1> <task2> ...",
		Short: "Create code project session (20+ tasks)",
		Long: `Create code project session (20+ tasks).

Session naming:
  --name: Creates {directory}-{name} (e.g., llm-todo-auth-system)
  Auto:   Creates {directory}-{timestamp} (e.g., llm-todo-20260122-1430)

Tip: Use --name for large projects.`,
		Example: `  llmtodo code "Design" "Implement" "Test" --name auth-system
  llmtodo code "Task1" "Task2"  # Auto: llm-todo-20260122-1430`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := generateSessionID(name)
			isAutoNamed := name == ""

			// Only prompt if explicitly requested (--prompt flag)
			// Default behavior: quick creation, enrich later via file
			// Note: skipPrompt=false means "do prompt", so we check for that
			if !skipPrompt {
				// Default: skip prompts (quick creation)
			} else {
				// User passed --prompt: do interactive prompts
				goal, boundaries, successCriteria = promptCodeSession()
			}

			err := createSessionWithMetadata(sessionID, "code", args, goal, boundaries, successCriteria, "")
			if err != nil {
				return err
			}

			// Set as current session
			if err := todo.SetCurrentSession(sessionID); err != nil {
				return err
			}

			// Show naming hint if auto-generated
			if isAutoNamed {
				printNamingHint()
			}

			// Show suggestion if goal still not provided
			if goal == "" {
				printGoalSuggestion(sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Session context (why these tasks exist)")
	cmd.Flags().StringVar(&boundaries, "boundaries", "", "What's out of scope")
	cmd.Flags().StringVar(&successCriteria, "success", "", "Success criteria")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name (creates {dir}-{name}, otherwise auto-timestamp)")
	cmd.Flags().BoolVar(&skipPrompt, "prompt", false, "Prompt for goal/boundaries/success interactively")

	return cmd
}

func researchCmd() *cobra.Command {
	var goal, deliverables, name string
	var skipPrompt bool

	cmd := &cobra.Command{
		Use:   "research <task1> <task2> ...",
		Short: "Create research project session",
		Long: `Create research project session.

Session naming:
  --name: Creates {directory}-{name} (e.g., llm-todo-enrichment-research)
  Auto:   Creates {directory}-{timestamp} (e.g., llm-todo-20260122-1430)`,
		Example: `  llmtodo research "Survey options" "Prototype" --name enrichment-research
  llmtodo research "Task1" "Task2"  # Auto: llm-todo-20260122-1430`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := generateSessionID(name)
			isAutoNamed := name == ""

			// Only prompt if explicitly requested (--prompt flag)
			// Default behavior: quick creation, enrich later via file
			if !skipPrompt {
				// Default: skip prompts (quick creation)
			} else {
				// User passed --prompt: do interactive prompts
				goal, deliverables = promptResearchSession()
			}

			err := createSessionWithMetadata(sessionID, "research", args, goal, "", "", deliverables)
			if err != nil {
				return err
			}

			// Set as current session
			if err := todo.SetCurrentSession(sessionID); err != nil {
				return err
			}

			// Show naming hint if auto-generated
			if isAutoNamed {
				printNamingHint()
			}

			// Show suggestion if goal still not provided
			if goal == "" {
				printGoalSuggestion(sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Session context (why these tasks exist)")
	cmd.Flags().StringVar(&deliverables, "deliverables", "", "Expected deliverables")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Session name (creates {dir}-{name}, otherwise auto-timestamp)")
	cmd.Flags().BoolVar(&skipPrompt, "prompt", false, "Prompt for goal/deliverables interactively")

	return cmd
}

func createSessionWithTasks(sessionID, sessionType string, tasks []string) error {
	return createSessionWithMetadata(sessionID, sessionType, tasks, "", "", "", "")
}

func createSessionWithMetadata(sessionID, sessionType string, tasks []string, goal, boundaries, successCriteria, deliverables string) error {
	mgr, err := todo.NewManager("")
	if err != nil {
		return err
	}
	defer mgr.Close()

	session, err := mgr.GetOrCreateSession(sessionID, sessionType)
	if err != nil {
		return err
	}

	// Update session metadata if provided
	updates := make(map[string]interface{})
	if goal != "" {
		updates["goal"] = goal
	}
	if boundaries != "" {
		updates["boundaries"] = encodeJSONArray(boundaries)
	}
	if successCriteria != "" {
		updates["success_criteria"] = successCriteria
	}
	if deliverables != "" {
		updates["deliverables"] = encodeJSONArray(deliverables)
	}

	if len(updates) > 0 {
		if err := mgr.UpdateSession(sessionID, updates); err != nil {
			return err
		}
	}

	fmt.Printf("Session: %s (%s)\n", session.ID, session.Type)
	if goal != "" {
		fmt.Printf("Goal: %s\n", goal)
	}
	if boundaries != "" {
		fmt.Printf("Boundaries: %s\n", boundaries)
	}
	if deliverables != "" {
		fmt.Printf("Deliverables: %s\n", deliverables)
	}

	fmt.Println()
	fmt.Printf("Created %d tasks:\n", len(tasks))

	// Track created task IDs and titles for enrichment file
	var createdTaskIDs []int64
	taskTitles := make(map[int64]string)

	for i, taskTitle := range tasks {
		task := &todo.Task{
			SessionID:     sessionID,
			Type:          "task",
			Priority:      "p0",
			PriorityOrder: (i + 1) * 100,
			Status:        "pending",
			Task:          taskTitle,
			ActiveForm:    generateActiveForm(taskTitle),
		}

		id, err := mgr.CreateTask(task)
		if err != nil {
			return err
		}
		fmt.Printf("  %d. %s\n", id, taskTitle)

		createdTaskIDs = append(createdTaskIDs, id)
		taskTitles[id] = taskTitle
	}

	// Generate enrichment file
	enrichmentPath, err := exporter.GetEnrichmentPath(sessionID)
	if err != nil {
		fmt.Printf("\n⚠️  Warning: Could not determine enrichment file path: %v\n", err)
		return nil
	}

	if err := exporter.GenerateEnrichmentFile(session, createdTaskIDs, taskTitles, enrichmentPath); err != nil {
		// Non-fatal: just warn and continue
		fmt.Printf("\n⚠️  Warning: Could not generate enrichment file: %v\n", err)
	} else {
		// Show recommendation
		fmt.Printf("\nPre-filled enrichment: %s\n\n", enrichmentPath)
		fmt.Println("RECOMMEND: Enrich tasks for better context")
		fmt.Println("  LLM workflow:")
		fmt.Printf("    1. Read: %s\n", enrichmentPath)
		fmt.Println("    2. Write: Overwrite entire file with enriched version")
		fmt.Printf("    3. Import: llmtodo import %s\n\n", enrichmentPath)
		fmt.Println("Skip enrichment: llmtodo next")
	}

	return nil
}
