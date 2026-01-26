package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBulkImportAtomic(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create valid YAML file with 10 tasks
	yamlContent := `goal: "Test atomic bulk import"

tasks:
  - title: "Task 1"
    priority: p0
  - title: "Task 2"
    priority: p0
  - title: "Task 3"
    priority: p1
  - title: "Task 4"
    priority: p1
  - title: "Task 5"
    priority: p2
  - title: "Task 6"
    priority: p2
  - title: "Task 7"
    priority: p3
  - title: "Task 8"
    priority: p3
  - title: "Task 9"
    priority: p4
  - title: "Task 10"
    priority: p4
`

	yamlFile := filepath.Join(s.dir, "tasks.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Import all tasks
	importOut, err := s.run("import", yamlFile)
	if err != nil {
		t.Fatalf("Import failed: %v\n%s", err, importOut)
	}

	if !strings.Contains(importOut, "Imported 10 tasks") {
		t.Errorf("Expected 10 tasks imported, got: %s", importOut)
	}

	// Verify goal was set (shown in import output)
	if !strings.Contains(importOut, "Goal: Test atomic bulk import") {
		t.Errorf("Expected goal to be set in import output, got: %s", importOut)
	}

	// Verify all tasks exist
	out, err := s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "10 total") {
		t.Errorf("Expected 10 tasks after import, got: %s", out)
	}
}

func TestBulkImportWithInvalidTask(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create YAML with invalid priority on task 5
	yamlContent := `tasks:
  - title: "Task 1"
    priority: p0
  - title: "Task 2"
    priority: p0
  - title: "Task 3"
    priority: p0
  - title: "Task 4"
    priority: p0
  - title: "Task 5"
    priority: INVALID
  - title: "Task 6"
    priority: p0
`

	yamlFile := filepath.Join(s.dir, "invalid.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Import should fail due to validation
	_, err := s.run("import", yamlFile)
	if err == nil {
		t.Error("Expected import to fail with invalid priority, but it succeeded")
	}

	// Verify NO tasks were created (atomic failure)
	out, err := s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "0 total") && !strings.Contains(out, "No tasks found") {
		t.Errorf("Expected 0 tasks after failed import (atomic rollback), got: %s", out)
	}
}

func TestBulkImportWithDependencies(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create YAML with task dependencies
	yamlContent := `tasks:
  - id: task-1
    title: "Database setup"
    priority: p0
  - id: task-2
    title: "API implementation"
    priority: p1
    depends_on:
      - task-1
  - id: task-3
    title: "Frontend integration"
    priority: p2
    depends_on:
      - task-1
      - task-2
`

	yamlFile := filepath.Join(s.dir, "deps.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	// Import with dependencies
	out, err := s.run("import", yamlFile)
	if err != nil {
		t.Fatalf("Import failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Imported 3 tasks") {
		t.Errorf("Expected 3 tasks imported, got: %s", out)
	}

	// The import creates 3 tasks - check all of them to find which has dependencies
	// Try tasks 1-10 to find the one with "Frontend integration"
	var foundTask bool
	for taskID := 1; taskID <= 10; taskID++ {
		out, err = s.run("show", fmt.Sprintf("%d", taskID))
		if err != nil {
			continue // Task doesn't exist
		}
		if strings.Contains(out, "Frontend integration") {
			t.Logf("Found Frontend integration at task %d:\n%s", taskID, out)
			foundTask = true
			break
		}
	}

	if !foundTask {
		t.Fatal("Could not find Frontend integration task")
	}
	if err != nil {
		t.Fatalf("show failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Dependencies") {
		t.Errorf("Expected dependencies to be shown, got: %s", out)
	}
}

func TestImportAndDelete(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Import tasks
	yamlContent := `tasks:
  - title: "Keep task 1"
    priority: p0
  - title: "Delete task 2"
    priority: p0
  - title: "Keep task 3"
    priority: p0
  - title: "Delete task 4"
    priority: p0
  - title: "Keep task 5"
    priority: p0
`

	yamlFile := filepath.Join(s.dir, "tasks.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}

	out, err := s.run("import", yamlFile)
	if err != nil {
		t.Fatalf("Import failed: %v\n%s", err, out)
	}

	// Delete some tasks
	out, err = s.run("delete", "2,4")
	if err != nil {
		t.Fatalf("Delete failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Deleted 2 tasks") {
		t.Errorf("Expected 2 tasks deleted, got: %s", out)
	}

	// Verify only 3 remain
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "3 total") {
		t.Errorf("Expected 3 tasks remaining, got: %s", out)
	}

	// Verify task 2 is gone
	_, err = s.run("show", "2")
	if err == nil {
		t.Error("Expected task 2 to be deleted, but it still exists")
	}

	// Verify task 1 still exists
	out, err = s.run("show", "1")
	if err != nil {
		t.Fatalf("Expected task 1 to exist: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Keep task 1") {
		t.Errorf("Task 1 content changed, got: %s", out)
	}
}
