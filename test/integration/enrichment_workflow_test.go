package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dtnitsch/llm-todo/internal/exporter"
	"github.com/dtnitsch/llm-todo/internal/importer"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

// TestEnrichmentWorkflow validates the complete enrichment workflow:
// 1. Create tasks (simulates quick/code/research commands)
// 2. Generate enrichment file
// 3. Update enrichment file (simulate LLM editing)
// 4. Re-import and verify updates
func TestEnrichmentWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		sessionType string
		taskTitles  []string
		sessionGoal string
		enrichments map[int64]struct {
			effort       string
			files        []string
			instructions map[string][]string
		}
		wantUpdated int
	}{
		{
			name:        "quick session with 3 tasks",
			sessionType: "quick",
			taskTitles:  []string{"Fix bug", "Write test", "Deploy"},
			sessionGoal: "Quick fixes for release",
			enrichments: map[int64]struct {
				effort       string
				files        []string
				instructions map[string][]string
			}{
				0: { // First task
					effort: "s",
					files:  []string{"src/main.go"},
					instructions: map[string][]string{
						"must_do":     {"Add validation"},
						"must_not_do": {"Don't break API"},
					},
				},
			},
			wantUpdated: 3, // All tasks updated (YAML contains all tasks with title/priority)
		},
		{
			name:        "code session with metadata",
			sessionType: "code",
			taskTitles:  []string{"Design schema", "Implement", "Test"},
			sessionGoal: "Auth system",
			enrichments: map[int64]struct {
				effort       string
				files        []string
				instructions map[string][]string
			}{
				0: {
					effort: "m",
					files:  []string{"db/schema.sql", "models/user.go"},
					instructions: map[string][]string{
						"must_do": {"Add indexes", "Add constraints"},
					},
				},
				1: {
					effort: "m",
					files:  []string{"handlers/auth.go"},
				},
			},
			wantUpdated: 3, // All tasks are updated (LLM rewrites entire file)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create test database
			tmpDir := t.TempDir()
			dbPath := filepath.Join(tmpDir, "test.db")
			os.Setenv("TODO_DB", dbPath)
			defer os.Unsetenv("TODO_DB")

			mgr, err := todo.NewManager(dbPath)
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}
			defer mgr.Close()

			// Step 1: Create session and tasks (simulates quick/code/research command)
			sessionID := "test-" + tt.name
			session, err := mgr.GetOrCreateSession(sessionID, tt.sessionType)
			if err != nil {
				t.Fatalf("Failed to create session: %v", err)
			}

			// Update session goal
			if tt.sessionGoal != "" {
				updates := map[string]interface{}{"goal": tt.sessionGoal}
				if err := mgr.UpdateSession(sessionID, updates); err != nil {
					t.Fatalf("Failed to update session goal: %v", err)
				}
			}

			// Create tasks
			var taskIDs []int64
			taskTitlesMap := make(map[int64]string)
			for i, title := range tt.taskTitles {
				task := &todo.Task{
					SessionID:     sessionID,
					Type:          "task",
					Priority:      "p0",
					PriorityOrder: (i + 1) * 100,
					Status:        "pending",
					Task:          title,
					ActiveForm:    "Working on: " + title,
				}

				id, err := mgr.CreateTask(task)
				if err != nil {
					t.Fatalf("Failed to create task: %v", err)
				}

				taskIDs = append(taskIDs, id)
				taskTitlesMap[id] = title
			}

			// Reload session to get updated metadata
			session, err = mgr.GetSession(sessionID)
			if err != nil {
				t.Fatalf("Failed to reload session: %v", err)
			}

			// Step 2: Generate enrichment file
			enrichmentPath := filepath.Join(tmpDir, "enrichment.yaml")
			err = exporter.GenerateEnrichmentFile(session, taskIDs, taskTitlesMap, enrichmentPath)
			if err != nil {
				t.Fatalf("Failed to generate enrichment file: %v", err)
			}

			// Verify file exists
			if _, err := os.Stat(enrichmentPath); os.IsNotExist(err) {
				t.Fatal("Enrichment file was not created")
			}

			// Step 3: Simulate editing enrichment file
			// (In real workflow, LLM would edit the file)
			// For test, we'll create a modified version
			enrichedPath := filepath.Join(tmpDir, "enriched.yaml")
			err = createEnrichedFile(enrichedPath, sessionID, tt.sessionGoal, taskIDs, taskTitlesMap, tt.enrichments)
			if err != nil {
				t.Fatalf("Failed to create enriched file: %v", err)
			}

			// Step 4: Re-import enriched file (update mode)
			count, _, err := importer.UpdateTasksFromYAML(mgr, sessionID, enrichedPath)
			if err != nil {
				t.Fatalf("Failed to update tasks from enrichment: %v", err)
			}

			if count != tt.wantUpdated {
				t.Errorf("Updated %d tasks, want %d", count, tt.wantUpdated)
			}

			// Step 5: Verify updates were applied
			for idx, enrichment := range tt.enrichments {
				if int(idx) >= len(taskIDs) {
					continue
				}

				taskID := int(taskIDs[idx])
				task, err := mgr.GetTask(taskID)
				if err != nil {
					t.Errorf("Failed to get task %d: %v", taskID, err)
					continue
				}

				// Verify effort
				if enrichment.effort != "" && task.Effort != enrichment.effort {
					t.Errorf("Task %d effort = %q, want %q", taskID, task.Effort, enrichment.effort)
				}

				// Verify files were set
				if len(enrichment.files) > 0 && task.Files == "[]" {
					t.Errorf("Task %d files not updated", taskID)
				}

				// Verify instructions were set
				if len(enrichment.instructions) > 0 && task.Instructions == "" {
					t.Errorf("Task %d instructions not updated", taskID)
				}
			}
		})
	}
}

// createEnrichedFile simulates LLM editing the enrichment file
func createEnrichedFile(path, sessionID, goal string, taskIDs []int64, taskTitles map[int64]string, enrichments map[int64]struct {
	effort       string
	files        []string
	instructions map[string][]string
}) error {
	content := "# Enriched file\n\n"

	if goal != "" {
		content += "goal: \"" + goal + "\"\n\n"
	}

	content += "tasks:\n"

	for idx, taskID := range taskIDs {
		content += fmt.Sprintf("  - id: task-%d\n", taskID)
		content += "    title: \"" + taskTitles[taskID] + "\"\n"
		content += "    priority: p0\n"

		// Add enrichment if provided
		if enrichment, ok := enrichments[int64(idx)]; ok {
			if enrichment.effort != "" {
				content += "    effort: \"" + enrichment.effort + "\"\n"
			}

			if len(enrichment.files) > 0 {
				content += "    files:\n"
				for _, file := range enrichment.files {
					content += "      - \"" + file + "\"\n"
				}
			}

			if len(enrichment.instructions) > 0 {
				content += "    instructions:\n"
				if mustDo, ok := enrichment.instructions["must_do"]; ok {
					content += "      must_do:\n"
					for _, item := range mustDo {
						content += "        - \"" + item + "\"\n"
					}
				}
				if mustNotDo, ok := enrichment.instructions["must_not_do"]; ok {
					content += "      must_not_do:\n"
					for _, item := range mustNotDo {
						content += "        - \"" + item + "\"\n"
					}
				}
			}
		}

		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}
