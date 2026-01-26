package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var todoBin string

func TestMain(m *testing.M) {
	// Build binary before tests
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	binPath := filepath.Join(projectRoot, "bin", "todo-test")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/todo")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("failed to build todo binary: " + err.Error() + "\n" + string(output))
	}
	todoBin = binPath

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(todoBin)
	os.Exit(code)
}

// Helper to run todo command in a temp directory
func runTodo(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Create temp dir for this test
	tempDir, err := os.MkdirTemp("", "todo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Use local database
	dbPath := filepath.Join(tempDir, ".llm-todo", "tasks.db")

	cmd := exec.Command(todoBin, args...)
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TODO_DB="+dbPath)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper with persistent session across commands
type Session struct {
	t       *testing.T
	dir     string
	dbPath  string
}

func newSession(t *testing.T) *Session {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "todo-session-*")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tempDir, ".llm-todo", "tasks.db")

	return &Session{
		t:      t,
		dir:    tempDir,
		dbPath: dbPath,
	}
}

func (s *Session) run(args ...string) (string, error) {
	s.t.Helper()

	// Use directory name as session ID
	sessionID := filepath.Base(s.dir)

	cmd := exec.Command(todoBin, args...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(),
		"TODO_DB="+s.dbPath,
		"TODO_SESSION="+sessionID,
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (s *Session) cleanup() {
	os.RemoveAll(s.dir)
}

func TestQuickWorkflow(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create quick session
	out, err := s.run("quick", "Fix bug", "Update docs", "Run tests")
	if err != nil {
		t.Fatalf("quick failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "✓ Session") {
		t.Errorf("Expected session created, got: %s", out)
	}

	// Get p0 tasks
	out, err = s.run("get", "p0")
	if err != nil {
		t.Fatalf("get p0 failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Fix bug") {
		t.Errorf("Expected 'Fix bug' in output, got: %s", out)
	}
	if !strings.Contains(out, "3 total") {
		t.Errorf("Expected 3 tasks, got: %s", out)
	}

	// Complete first task
	out, err = s.run("done", "1")
	if err != nil {
		t.Fatalf("done failed: %v\n%s", err, out)
	}

	// Check status
	out, err = s.run("status")
	if err != nil {
		t.Fatalf("status failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Completed: 1") {
		t.Errorf("Expected 1 completed, got: %s", out)
	}
	if !strings.Contains(out, "Pending: 2") {
		t.Errorf("Expected 2 pending, got: %s", out)
	}
}

func TestBatchOperations(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create session with 5 tasks
	s.run("quick", "Task 1", "Task 2", "Task 3", "Task 4", "Task 5")

	// Batch complete
	out, err := s.run("done", "1,2,3")
	if err != nil {
		t.Fatalf("batch done failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Completed 3 tasks") {
		t.Errorf("Expected 3 tasks completed, got: %s", out)
	}

	// Check remaining
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "2 total") {
		t.Errorf("Expected 2 pending tasks, got: %s", out)
	}

	// Batch block
	out, err = s.run("block", "4,5", "waiting on review")
	if err != nil {
		t.Fatalf("batch block failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Blocked 2 tasks") {
		t.Errorf("Expected 2 tasks blocked, got: %s", out)
	}
}

func TestShowCommand(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	s.run("quick", "Test task")

	out, err := s.run("show", "1")
	if err != nil {
		t.Fatalf("show failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Task #1") {
		t.Errorf("Expected task #1, got: %s", out)
	}
	if !strings.Contains(out, "Test task") {
		t.Errorf("Expected task title, got: %s", out)
	}
	if !strings.Contains(out, "Status:") {
		t.Errorf("Expected status, got: %s", out)
	}
}

func TestPriorityOrdering(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	s.run("quick", "Task 1", "Task 2", "Task 3")

	// Reorder task 3 to first position
	out, err := s.run("priority", "3", "50")
	if err != nil {
		t.Fatalf("priority failed: %v\n%s", err, out)
	}

	// Check order
	out, err = s.run("get", "p0")
	if err != nil {
		t.Fatalf("get p0 failed: %v\n%s", err, out)
	}

	// Task 3 should be first now (order 50 < 100)
	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got: %s", out)
	}

	// First task line should contain "Task 3"
	firstTask := lines[1] // Skip header
	if !strings.Contains(firstTask, "3") && !strings.Contains(firstTask, "Task 3") {
		t.Errorf("Expected task 3 first, got: %s", firstTask)
	}
}

func TestYAMLImport(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create test YAML file
	yamlContent := `- title: "Task from YAML"
  priority: p0
  status: pending
  effort: s

- title: "Another task"
  priority: p1
  status: pending
`

	yamlPath := filepath.Join(s.dir, "test.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Import
	out, err := s.run("import", yamlPath)
	if err != nil {
		t.Fatalf("import failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Imported 2 tasks") {
		t.Errorf("Expected 2 tasks imported, got: %s", out)
	}

	// Verify import
	out, err = s.run("get", "p0")
	if err != nil {
		t.Fatalf("get p0 failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "Task from YAML") {
		t.Errorf("Expected imported task, got: %s", out)
	}
}

func TestNextCommand(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	s.run("quick", "First task", "Second task")

	out, err := s.run("next")
	if err != nil {
		t.Fatalf("next failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "🎯 NEXT:") {
		t.Errorf("Expected NEXT header, got: %s", out)
	}
	if !strings.Contains(out, "First task") {
		t.Errorf("Expected first task, got: %s", out)
	}
	if !strings.Contains(out, "task 1/2") {
		t.Errorf("Expected task counter, got: %s", out)
	}
}

func TestEmptySession(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Try to get tasks from empty session
	out, err := s.run("get", "p0")

	// Should not error, just show no tasks
	if err != nil {
		t.Fatalf("get p0 on empty session failed: %v\n%s", err, out)
	}

	if !strings.Contains(out, "No tasks found") && !strings.Contains(out, "0 total") {
		t.Logf("Expected empty tasks message, got: %s", out)
	}
}

func TestInvalidTaskID(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	s.run("quick", "Task 1")

	// Try to show non-existent task
	_, err := s.run("show", "999")
	if err == nil {
		t.Error("Expected error for invalid task ID, got none")
	}
}

func TestGetFilters(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	s.run("quick", "Task 1", "Task 2", "Task 3")
	s.run("done", "1")
	s.run("block", "2", "blocked")

	tests := []struct {
		filter   string
		expected string
	}{
		{"pending", "1 total"},
		{"completed", "1 total"},
		{"blocked", "1 total"},
		{"p0", "3 total"},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			out, err := s.run("get", tt.filter)
			if err != nil {
				t.Fatalf("get %s failed: %v\n%s", tt.filter, err, out)
			}

			if !strings.Contains(out, tt.expected) {
				t.Errorf("Filter %s: expected %s, got: %s", tt.filter, tt.expected, out)
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	s := newSession(t)
	defer s.cleanup()

	// Create 5 tasks
	s.run("quick", "Task 1", "Task 2", "Task 3", "Task 4", "Task 5")

	// Verify all tasks exist
	out, err := s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "5 total") {
		t.Errorf("Expected 5 tasks, got: %s", out)
	}

	// Delete single task
	out, err = s.run("delete", "3")
	if err != nil {
		t.Fatalf("delete failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Deleted task #3") {
		t.Errorf("Expected delete confirmation, got: %s", out)
	}

	// Verify task is gone
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "4 total") {
		t.Errorf("Expected 4 tasks after delete, got: %s", out)
	}

	// Batch delete
	out, err = s.run("delete", "1,2")
	if err != nil {
		t.Fatalf("batch delete failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Deleted 2 tasks") {
		t.Errorf("Expected batch delete confirmation, got: %s", out)
	}

	// Verify only 2 tasks remain
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "2 total") {
		t.Errorf("Expected 2 tasks after batch delete, got: %s", out)
	}

	// Test rm alias
	out, err = s.run("rm", "4")
	if err != nil {
		t.Fatalf("rm alias failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Deleted task #4") {
		t.Errorf("Expected rm alias to work, got: %s", out)
	}

	// Test remove alias
	out, err = s.run("remove", "5")
	if err != nil {
		t.Fatalf("remove alias failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Deleted task #5") {
		t.Errorf("Expected remove alias to work, got: %s", out)
	}

	// Verify all tasks deleted
	out, err = s.run("get", "pending")
	if err != nil {
		t.Fatalf("get pending failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "0 total") {
		t.Errorf("Expected 0 tasks after all deletes, got: %s", out)
	}

	// Test delete non-existent task
	_, err = s.run("delete", "999")
	if err == nil {
		t.Error("Expected error when deleting non-existent task, got none")
	}
}
