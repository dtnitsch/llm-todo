package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var todoBin string

func init() {
	// Build binary for integration tests
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(wd))
	binPath := filepath.Join(projectRoot, "bin", "todo-integration-test")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/llmtodo")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("failed to build llmtodo binary for integration tests: " + err.Error() + "\n" + string(output))
	}
	todoBin = binPath
}

// Session represents a test session with persistent state
type Session struct {
	t      *testing.T
	dir    string
	dbPath string
}

// newSession creates a new test session
func newSession(t *testing.T) *Session {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "llmtodo-integration-*")
	if err != nil {
		t.Fatal(err)
	}

	// Create .llm-todo directory for project-local mode
	llmTodoDir := filepath.Join(tempDir, ".llm-todo")
	if err := os.MkdirAll(llmTodoDir, 0755); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(llmTodoDir, "tasks.db")

	return &Session{
		t:      t,
		dir:    tempDir,
		dbPath: dbPath,
	}
}

// run executes a command in the session's context
func (s *Session) run(args ...string) (string, error) {
	s.t.Helper()

	cmd := exec.Command(todoBin, args...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(),
		"TODO_DB="+s.dbPath,
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// cleanup removes the session directory
func (s *Session) cleanup() {
	os.RemoveAll(s.dir)
}
