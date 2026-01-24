package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

func TestGenerateEnrichmentFile(t *testing.T) {
	tests := []struct {
		name           string
		session        *todo.Session
		taskIDs        []int64
		taskTitles     map[int64]string
		wantGoal       bool
		wantSuccess    bool
		wantBoundaries bool
		wantTaskCount  int
	}{
		{
			name: "basic session with goal",
			session: &todo.Session{
				ID:   "test-session-001",
				Type: "code",
				Goal: "Implement feature X",
			},
			taskIDs: []int64{1, 2, 3},
			taskTitles: map[int64]string{
				1: "Task one",
				2: "Task two",
				3: "Task three",
			},
			wantGoal:      true,
			wantTaskCount: 3,
		},
		{
			name: "session with all metadata",
			session: &todo.Session{
				ID:              "test-session-002",
				Type:            "code",
				Goal:            "Build auth system",
				SuccessCriteria: "All tests pass",
				Boundaries:      `["No OAuth", "Local only"]`,
			},
			taskIDs: []int64{10, 20},
			taskTitles: map[int64]string{
				10: "Design schema",
				20: "Write tests",
			},
			wantGoal:       true,
			wantSuccess:    true,
			wantBoundaries: true,
			wantTaskCount:  2,
		},
		{
			name: "session without metadata",
			session: &todo.Session{
				ID:   "test-session-003",
				Type: "quick",
			},
			taskIDs: []int64{5},
			taskTitles: map[int64]string{
				5: "Single task",
			},
			wantGoal:      false,
			wantTaskCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "enrichment.yaml")

			// Generate enrichment file
			err := GenerateEnrichmentFile(tt.session, tt.taskIDs, tt.taskTitles, filePath)
			if err != nil {
				t.Fatalf("GenerateEnrichmentFile() error = %v", err)
			}

			// Read generated file
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			// Verify header comment
			if !strings.Contains(contentStr, "# Auto-generated enrichment file") {
				t.Error("Missing header comment")
			}

			if !strings.Contains(contentStr, "DO NOT EDIT IN PLACE") {
				t.Error("Missing LLM instruction comment")
			}

			if !strings.Contains(contentStr, tt.session.ID) {
				t.Errorf("Missing session ID %s in header", tt.session.ID)
			}

			// Verify goal
			if tt.wantGoal {
				if !strings.Contains(contentStr, "goal:") {
					t.Error("Missing goal field")
				}
				if !strings.Contains(contentStr, tt.session.Goal) {
					t.Errorf("Goal not in file. Want %q", tt.session.Goal)
				}
			}

			// Verify success criteria
			if tt.wantSuccess {
				if !strings.Contains(contentStr, "success_criteria:") {
					t.Error("Missing success_criteria field")
				}
				if !strings.Contains(contentStr, tt.session.SuccessCriteria) {
					t.Errorf("SuccessCriteria not in file. Want %q", tt.session.SuccessCriteria)
				}
			}

			// Verify boundaries
			if tt.wantBoundaries {
				if !strings.Contains(contentStr, "boundaries:") {
					t.Error("Missing boundaries field")
				}
			}

			// Verify example task section
			if !strings.Contains(contentStr, "EXAMPLE TASK") {
				t.Error("Missing example task")
			}

			if !strings.Contains(contentStr, "id: example-task") {
				t.Error("Missing example task ID")
			}

			if !strings.Contains(contentStr, "REMOVE THIS FROM YOUR OUTPUT") {
				t.Error("Missing removal instruction for example task")
			}

			// Verify example shows all fields
			if !strings.Contains(contentStr, "# p0 (critical)") {
				t.Error("Missing priority explanation in example")
			}

			if !strings.Contains(contentStr, "# xs, s, m") {
				t.Error("Missing effort explanation in example")
			}

			if !strings.Contains(contentStr, "must_do:") {
				t.Error("Missing must_do in example")
			}

			if !strings.Contains(contentStr, "must_not_do:") {
				t.Error("Missing must_not_do in example")
			}

			// Verify tasks section
			if !strings.Contains(contentStr, "tasks:") {
				t.Error("Missing tasks section")
			}

			if !strings.Contains(contentStr, "# Your actual tasks") {
				t.Error("Missing actual tasks section header")
			}

			// Verify each task is MINIMAL (just id and title)
			for taskID, title := range tt.taskTitles {
				if !strings.Contains(contentStr, "task-") {
					t.Errorf("Missing task ID format for task %d", taskID)
				}

				if !strings.Contains(contentStr, title) {
					t.Errorf("Missing task title %q", title)
				}
			}
		})
	}
}

func TestGetEnrichmentPath(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{
			name:      "valid session ID",
			sessionID: "llm-todo-20260124-1234",
			wantErr:   false,
		},
		{
			name:      "session with special chars",
			sessionID: "my-project-feature-x",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetEnrichmentPath(tt.sessionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEnrichmentPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify path format
				if !strings.Contains(path, ".llm-todo/enrichment") {
					t.Errorf("Path doesn't contain .llm-todo/enrichment: %s", path)
				}

				if !strings.HasSuffix(path, tt.sessionID+".yaml") {
					t.Errorf("Path doesn't end with %s.yaml: %s", tt.sessionID, path)
				}

				// Verify directory was created
				dir := filepath.Dir(path)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("Directory was not created: %s", dir)
				}
			}
		})
	}
}

func TestEscapeYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no quotes",
			input: "simple string",
			want:  "simple string",
		},
		{
			name:  "with quotes",
			input: `say "hello"`,
			want:  `say \"hello\"`,
		},
		{
			name:  "multiple quotes",
			input: `"quoted" and "more"`,
			want:  `\"quoted\" and \"more\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeYAML(tt.input)
			if got != tt.want {
				t.Errorf("escapeYAML() = %q, want %q", got, tt.want)
			}
		})
	}
}
