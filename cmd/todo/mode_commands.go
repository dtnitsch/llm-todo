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
	return &cobra.Command{
		Use:   "quick <task1> <task2> ...",
		Short: "Create quick session (3-5 tasks)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())
			return createSessionWithTasks(sessionID, "quick", args)
		},
	}
}

func codeCmd() *cobra.Command{
	return &cobra.Command{
		Use:   "code <task1> <task2> ...",
		Short: "Create code project session (20+ tasks)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())

			// Prompt for code session metadata
			goal, boundaries, successCriteria := promptCodeSession()

			return createSessionWithMetadata(sessionID, "code", args, goal, boundaries, successCriteria, "")
		},
	}
}

func researchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "research <task1> <task2> ...",
		Short: "Create research project session",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := filepath.Base(mustGetwd())

			// Prompt for research session metadata
			goal, deliverables := promptResearchSession()

			return createSessionWithMetadata(sessionID, "research", args, goal, "", "", deliverables)
		},
	}
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

	fmt.Printf("✓ Session: %s (%s)\n", session.ID, session.Type)
	if goal != "" {
		fmt.Printf("  Goal: %s\n", goal)
	}
	if boundaries != "" {
		fmt.Printf("  Boundaries: %s\n", boundaries)
	}
	if deliverables != "" {
		fmt.Printf("  Deliverables: %s\n", deliverables)
	}

	fmt.Println()
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

	fmt.Printf("\nNext: todo next\n")
	return nil
}
