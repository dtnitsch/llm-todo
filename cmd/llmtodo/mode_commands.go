package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addModeCommands(root *cobra.Command) {
	root.AddCommand(quickCmd())
	root.AddCommand(codeCmd())
	root.AddCommand(researchCmd())
}

func quickCmd() *cobra.Command {
	var goal string

	cmd := &cobra.Command{
		Use:   "quick <task1> <task2> ...",
		Short: "Create quick session (3-5 tasks)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())
			err := createSessionWithMetadata(sessionID, "quick", args, goal, "", "", "")
			if err != nil {
				return err
			}

			// Show suggestion if goal not provided
			if goal == "" {
				printGoalSuggestion(sessionID)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&goal, "goal", "g", "", "Session context (why these tasks exist)")
	return cmd
}

func codeCmd() *cobra.Command{
	var goal, boundaries, successCriteria string
	var skipPrompt bool

	cmd := &cobra.Command{
		Use:   "code <task1> <task2> ...",
		Short: "Create code project session (20+ tasks)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())

			// If flags not provided and not skipping prompt, prompt interactively
			if !skipPrompt && goal == "" {
				goal, boundaries, successCriteria = promptCodeSession()
			}

			err := createSessionWithMetadata(sessionID, "code", args, goal, boundaries, successCriteria, "")
			if err != nil {
				return err
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
	cmd.Flags().BoolVar(&skipPrompt, "skip-prompt", false, "Skip interactive prompts")

	return cmd
}

func researchCmd() *cobra.Command {
	var goal, deliverables string
	var skipPrompt bool

	cmd := &cobra.Command{
		Use:   "research <task1> <task2> ...",
		Short: "Create research project session",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())

			// If flags not provided and not skipping prompt, prompt interactively
			if !skipPrompt && goal == "" {
				goal, deliverables = promptResearchSession()
			}

			err := createSessionWithMetadata(sessionID, "research", args, goal, "", "", deliverables)
			if err != nil {
				return err
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
	cmd.Flags().BoolVar(&skipPrompt, "skip-prompt", false, "Skip interactive prompts")

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
	}

	fmt.Printf("\nRun: llmtodo next\n")
	return nil
}
