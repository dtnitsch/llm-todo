package main

import (
	"fmt"

	"github.com/dtnitsch/llm-todo/internal/todo"
	"github.com/dtnitsch/llm-todo/internal/validation"
)

func checkUnblockedTasks(mgr *todo.Manager, sessionID string, completedID int) error {
	unblocked, err := validation.FindUnblockedTasks(mgr, sessionID, completedID)
	if err != nil {
		return err
	}

	if len(unblocked) > 0 {
		fmt.Printf("\n📌 Unblocked tasks:\n")
		for _, task := range unblocked {
			fmt.Printf("  #%d: %s\n", task.ID, task.Task)
		}
	}

	return nil
}
