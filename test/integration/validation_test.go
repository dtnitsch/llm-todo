package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFlag(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	tests := []struct {
		name        string
		yamlContent string
		wantError   bool
		wantMessage string
	}{
		{
			name: "valid YAML",
			yamlContent: `goal: "Test validation"

tasks:
  - title: "Task 1"
    priority: p0
    effort: m
  - title: "Task 2"
    priority: p1
    effort: s
`,
			wantError:   false,
			wantMessage: "Validation PASSED",
		},
		{
			name: "typo in field name",
			yamlContent: `tasks:
  - title: "Task 1"
    priortiy: p0
`,
			wantError:   true,
			wantMessage: "Found 'priortiy' - did you mean 'priority'",
		},
		{
			name: "invalid priority",
			yamlContent: `tasks:
  - title: "Task 1"
    priority: P0
`,
			wantError:   true,
			wantMessage: "Invalid priority",
		},
		{
			name: "invalid effort",
			yamlContent: `tasks:
  - title: "Task 1"
    effort: large
`,
			wantError:   true,
			wantMessage: "Invalid effort",
		},
		{
			name: "invalid task type",
			yamlContent: `tasks:
  - title: "Task 1"
    type: invalid-type
`,
			wantError:   true,
			wantMessage: "Invalid task type",
		},
		{
			name: "malformed YAML",
			yamlContent: `tasks:
  - title: "Task 1
    priority: p0
`,
			wantError:   true,
			wantMessage: "Invalid YAML format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlFile := filepath.Join(s.dir, tt.name+".yaml")
			if err := os.WriteFile(yamlFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to write YAML: %v", err)
			}

			out, err := s.run("import", "--validate", yamlFile)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected validation to fail, but it passed: %s", out)
				}
				if !strings.Contains(out, tt.wantMessage) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.wantMessage, out)
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to pass, but got error: %v\n%s", err, out)
				}
				if !strings.Contains(out, tt.wantMessage) {
					t.Errorf("Expected output to contain %q, got: %s", tt.wantMessage, out)
				}
			}

			// Verify no tasks were created during validation
			listOut, listErr := s.run("get", "pending")
			if listErr != nil {
				t.Fatalf("get pending failed: %v\n%s", listErr, listOut)
			}
			if !strings.Contains(listOut, "0 total") && !strings.Contains(listOut, "No tasks found") {
				t.Errorf("Expected 0 tasks after validation (no DB changes), got: %s", listOut)
			}
		})
	}
}

func TestValidateUpdateMode(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// First create some tasks
	s.run("quick", "Task 1", "Task 2", "Task 3")

	tests := []struct {
		name        string
		yamlContent string
		wantError   bool
		wantMode    string
	}{
		{
			name: "valid update YAML",
			yamlContent: `tasks:
  - id: task-1
    title: "Updated Task 1"
    priority: p0
    effort: m
`,
			wantError: false,
			wantMode:  "Update (enrichment file with task IDs)",
		},
		{
			name: "valid create YAML",
			yamlContent: `tasks:
  - title: "New Task"
    priority: p0
`,
			wantError: false,
			wantMode:  "Create (new tasks)",
		},
		{
			name: "invalid update YAML - bad priority",
			yamlContent: `tasks:
  - id: task-1
    title: "Task 1"
    priority: INVALID
`,
			wantError: true,
			wantMode:  "Update (enrichment file with task IDs)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlFile := filepath.Join(s.dir, tt.name+".yaml")
			if err := os.WriteFile(yamlFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to write YAML: %v", err)
			}

			out, err := s.run("import", "--validate", yamlFile)

			// Check mode detection
			if !strings.Contains(out, tt.wantMode) {
				t.Errorf("Expected mode %q in output, got: %s", tt.wantMode, out)
			}

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected validation to fail, but it passed: %s", out)
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to pass, but got error: %v\n%s", err, out)
				}
			}
		})
	}
}

func TestValidateNoDBAccess(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create valid YAML
	yamlContent := `tasks:
  - title: "Task 1"
    priority: p0
  - title: "Task 2"
    priority: p1
`

	yamlFile := filepath.Join(s.dir, "tasks.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Run validation
	out, err := s.run("import", "--validate", yamlFile)
	if err != nil {
		t.Fatalf("Validation failed: %v\n%s", err, out)
	}

	// Verify output shows what would be processed
	if !strings.Contains(out, "Would process: 2 tasks") {
		t.Errorf("Expected 'Would process: 2 tasks', got: %s", out)
	}

	// Verify NO database file was created or modified
	// Since we use --validate, it should be parse-only with no DB access
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "0 total") && !strings.Contains(out, "No tasks found") {
		t.Errorf("Expected 0 tasks (validation should not create tasks), got: %s", out)
	}
}

func TestValidateWithGoal(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	yamlContent := `goal: "Build authentication system"

tasks:
  - title: "Design schema"
    priority: p0
  - title: "Implement handlers"
    priority: p1
`

	yamlFile := filepath.Join(s.dir, "tasks.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	out, err := s.run("import", "--validate", yamlFile)
	if err != nil {
		t.Fatalf("Validation failed: %v\n%s", err, out)
	}

	// Verify goal is shown in validation output
	if !strings.Contains(out, "Build authentication system") {
		t.Errorf("Expected goal in validation output, got: %s", out)
	}

	if !strings.Contains(out, "Validation PASSED") {
		t.Errorf("Expected validation to pass, got: %s", out)
	}
}
