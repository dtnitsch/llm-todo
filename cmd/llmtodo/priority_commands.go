package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addPriorityCommands(root *cobra.Command) {
	root.AddCommand(priorityCmd())
}

func priorityCmd() *cobra.Command {
	var add bool

	cmd := &cobra.Command{
		Use:   "priority <task-ids> <order>",
		Short: "Set priority order for task(s)",
		Example: `  todo priority 4 50          # Set task 4 to order 50
  todo priority 3 100 --add   # Increase task 3 order by 100
  todo priority 2,3,7 25      # Set multiple to order 25`,
		Args: cobra.ExactArgs(2),
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

			order, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid order: %s", args[1])
			}

			if add {
				// Add to existing order
				for _, id := range ids {
					task, err := mgr.GetTask(id)
					if err != nil {
						return err
					}
					newOrder := task.PriorityOrder + order
					if err := mgr.UpdateTask(id, map[string]interface{}{"priority_order": newOrder}); err != nil {
						return err
					}
				}
			} else {
				// Set absolute order
				if err := mgr.BatchUpdatePriority(ids, order); err != nil {
					return err
				}
			}

			fmt.Printf("✓ Updated priority order for %d tasks\n", len(ids))
			return nil
		},
	}

	cmd.Flags().BoolVar(&add, "add", false, "Add to existing order instead of setting")

	return cmd
}
