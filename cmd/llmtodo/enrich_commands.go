package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func init() {
	rootCmd.AddCommand(enrichCmd())
}

func enrichCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enrich <task-id>",
		Short: "Add structured enrichment to tasks (non-blocking, LLM-friendly)",
		Long: `Add enrichment to tasks using flags. Designed for LLM workflows - no interactive prompts.

Enrichment types (0-5 scale):
  - instructions: must_do/must_not_do (concise, imperative, no prose)
  - files: file paths (no descriptions)
  - output: concrete deliverable (single sentence)
  - context: WHY task exists (1-2 sentences, present tense)
  - dependencies: prerequisite task IDs

LLM-friendly guidelines:
  - Concise > verbose
  - Imperative > descriptive
  - Factual > emotional
  - No prose, no emojis, no past-tense storytelling

Examples:
  # Add instructions (imperative, actionable)
  llmtodo enrich 1 --must-do "Add rate limiting middleware" "Return 429 on limit exceeded" --must-not "Skip health check endpoint"

  # Add files and output
  llmtodo enrich 1 --files "auth.go,auth_test.go" --output "Working auth endpoint with rate limiting"

  # Add context (WHY, present tense)
  llmtodo enrich 1 --notes "Validates enrichment provides enough context for cold-start LLMs" --deps "2,3"

  # Check enrichment status
  llmtodo enrich 1 --status

  # Get suggestions for what to add
  llmtodo enrich 1 --suggest

  # Enrich session goal
  llmtodo enrich --session --goal "Build user authentication"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			sessionID := getSessionID()

			// Handle session enrichment
			enrichSession, _ := cmd.Flags().GetBool("session")
			if enrichSession {
				return enrichSessionFromFlags(mgr, sessionID, cmd)
			}

			// Task enrichment requires task ID
			if len(args) == 0 {
				return fmt.Errorf("provide task ID (e.g., llmtodo enrich 1)")
			}

			taskID, err := parseTaskID(args[0])
			if err != nil {
				return err
			}

			task, err := mgr.GetTask(taskID)
			if err != nil {
				return err
			}

			// Handle status check
			showStatus, _ := cmd.Flags().GetBool("status")
			if showStatus {
				return showEnrichmentStatus(task)
			}

			// Handle suggestions
			showSuggest, _ := cmd.Flags().GetBool("suggest")
			if showSuggest {
				return showEnrichmentSuggestions(task)
			}

			// Apply enrichment from flags
			return applyEnrichment(mgr, task, cmd)
		},
	}

	// Enrichment flags
	cmd.Flags().StringSlice("must-do", []string{}, "Must-do instructions (repeatable)")
	cmd.Flags().StringSlice("must-not", []string{}, "Must-NOT-do instructions (repeatable)")
	cmd.Flags().String("files", "", "Comma-separated file paths")
	cmd.Flags().String("output", "", "Expected output/deliverable")
	cmd.Flags().String("notes", "", "Context/background notes")
	cmd.Flags().Bool("replace-notes", false, "Replace notes instead of appending")
	cmd.Flags().String("deps", "", "Comma-separated prerequisite task IDs")

	// Session enrichment
	cmd.Flags().Bool("session", false, "Enrich session instead of task")
	cmd.Flags().String("goal", "", "Session goal")
	cmd.Flags().String("boundaries", "", "Session boundaries (comma-separated)")
	cmd.Flags().String("success", "", "Success criteria")

	// Utility flags
	cmd.Flags().Bool("status", false, "Show enrichment status")
	cmd.Flags().Bool("suggest", false, "Show enrichment suggestions")

	return cmd
}

func applyEnrichment(mgr *todo.Manager, task *todo.Task, cmd *cobra.Command) error {
	updates := make(map[string]interface{})

	// Instructions
	mustDo, _ := cmd.Flags().GetStringSlice("must-do")
	mustNot, _ := cmd.Flags().GetStringSlice("must-not")

	if len(mustDo) > 0 || len(mustNot) > 0 {
		// Parse existing instructions
		existing := make(map[string][]string)
		if task.Instructions != "" && task.Instructions != "{}" {
			json.Unmarshal([]byte(task.Instructions), &existing)
		}

		// Merge with new instructions
		if len(mustDo) > 0 {
			existing["must_do"] = append(existing["must_do"], mustDo...)
		}
		if len(mustNot) > 0 {
			existing["must_not_do"] = append(existing["must_not_do"], mustNot...)
		}

		instructionsJSON, _ := json.Marshal(existing)
		updates["instructions"] = string(instructionsJSON)
	}

	// Files
	files, _ := cmd.Flags().GetString("files")
	if files != "" {
		updates["files"] = encodeJSONArray(files)
	}

	// Output
	output, _ := cmd.Flags().GetString("output")
	if output != "" {
		updates["output"] = output
	}

	// Notes
	notes, _ := cmd.Flags().GetString("notes")
	replaceNotes, _ := cmd.Flags().GetBool("replace-notes")
	if notes != "" {
		if replaceNotes || task.Notes == "" {
			updates["notes"] = notes
		} else {
			// Append to existing notes
			updates["notes"] = task.Notes + "\n" + notes
		}
	}

	// Dependencies
	deps, _ := cmd.Flags().GetString("deps")
	if deps != "" {
		updates["dependant_ids"] = encodeJSONArray(deps)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no enrichment provided. Use --must-do, --files, --output, --notes, or --deps")
	}

	// Apply updates
	if err := mgr.UpdateTask(task.ID, updates); err != nil {
		return err
	}

	// Show updated enrichment status
	updatedTask, _ := mgr.GetTask(task.ID)
	status := todo.GetEnrichmentStatus(updatedTask)

	fmt.Printf("✅ Task #%d enriched (%d/5 enrichments)\n", task.ID, status.Score())

	// Show hint if not fully enriched
	if hint := todo.GetEnrichmentHint(updatedTask); hint != "" {
		fmt.Println(hint)
	}

	return nil
}

func showEnrichmentStatus(task *todo.Task) error {
	status := todo.GetEnrichmentStatus(task)
	score := status.Score()

	fmt.Printf("Task #%d: %s\n", task.ID, task.Task)
	fmt.Printf("Enrichment: %d/5\n\n", score)

	printEnrichmentField("Instructions", status.HasInstructions, task.Instructions)
	printEnrichmentField("Files", status.HasFiles, task.Files)
	printEnrichmentField("Output", status.HasOutput, task.Output)
	printEnrichmentField("Context", status.HasContext, task.Notes)
	printEnrichmentField("Dependencies", status.HasDependencies, task.DependantIDs)

	if score < 5 {
		fmt.Printf("\n%s\n", todo.GetEnrichmentHint(task))
	}

	return nil
}

func showEnrichmentSuggestions(task *todo.Task) error {
	suggestions := todo.GetEnrichmentSuggestions(task)

	if len(suggestions) == 0 {
		fmt.Printf("✅ Task #%d is fully enriched (5/5)\n", task.ID)
		return nil
	}

	status := todo.GetEnrichmentStatus(task)
	fmt.Printf("Task #%d has %d/5 enrichments. Suggestions:\n\n", task.ID, status.Score())

	for i, s := range suggestions {
		fmt.Printf("%d. %s\n", i+1, s.Description)
		if s.Example != "" {
			fmt.Printf("   %s\n", s.Example)
		}
		fmt.Println()
	}

	return nil
}

func enrichSessionFromFlags(mgr *todo.Manager, sessionID string, cmd *cobra.Command) error {
	session, err := mgr.GetSession(sessionID)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	goal, _ := cmd.Flags().GetString("goal")
	if goal != "" {
		updates["goal"] = goal
	}

	boundaries, _ := cmd.Flags().GetString("boundaries")
	if boundaries != "" {
		updates["boundaries"] = encodeJSONArray(boundaries)
	}

	success, _ := cmd.Flags().GetString("success")
	if success != "" {
		updates["success_criteria"] = success
	}

	if len(updates) == 0 {
		return fmt.Errorf("no session enrichment provided. Use --goal, --boundaries, or --success")
	}

	if err := mgr.UpdateSession(sessionID, updates); err != nil {
		return err
	}

	fmt.Printf("✅ Session '%s' enriched\n", session.ID)
	return nil
}

func printEnrichmentField(name string, hasValue bool, value string) {
	status := "❌"
	if hasValue {
		status = "✅"
	}

	fmt.Printf("%s %s", status, name)

	if hasValue && value != "" {
		// Truncate long values
		displayValue := value
		if len(displayValue) > 60 {
			displayValue = displayValue[:57] + "..."
		}
		fmt.Printf(": %s", displayValue)
	}

	fmt.Println()
}

func parseTaskID(arg string) (int, error) {
	ids, err := todo.ParseTaskIDs(arg)
	if err != nil {
		return 0, err
	}
	if len(ids) != 1 {
		return 0, fmt.Errorf("provide exactly one task ID")
	}
	return ids[0], nil
}
