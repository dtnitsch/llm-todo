package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// TaskExport represents a task in export format
type TaskExport struct {
	ID           string                 `yaml:"id,omitempty"`
	Title        string                 `yaml:"title"`
	Type         string                 `yaml:"type,omitempty"`
	Priority     string                 `yaml:"priority,omitempty"`
	Status       string                 `yaml:"status,omitempty"`
	Effort       string                 `yaml:"effort,omitempty"`
	Files        []string               `yaml:"files,omitempty"`
	Instructions map[string][]string    `yaml:"instructions,omitempty"`
	Refs         []string               `yaml:"refs,omitempty"`
	DependsOn    []string               `yaml:"depends_on,omitempty"`
}

// FileExport represents a YAML file with goal and tasks
type FileExport struct {
	Goal  string       `yaml:"goal,omitempty"`
	Tasks []TaskExport `yaml:"tasks,omitempty"`
}

// ExportToYAML exports tasks to a single YAML file
func ExportToYAML(mgr *todo.Manager, sessionID string, filePath string) error {
	session, err := mgr.GetSession(sessionID)
	if err != nil {
		return err
	}

	tasks, err := mgr.ListTasks(sessionID, nil)
	if err != nil {
		return err
	}

	// Convert to export format
	var exports []TaskExport
	for _, task := range tasks {
		exports = append(exports, convertTaskToExport(task))
	}

	// Create file export with goal
	fileExport := FileExport{
		Goal:  session.Goal,
		Tasks: exports,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(fileExport)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported %d tasks to: %s\n", len(exports), filePath)
	return nil
}

// ExportByPriority exports tasks split by priority
func ExportByPriority(mgr *todo.Manager, sessionID string, dirPath string, includeDone bool) error {
	// Create directory
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get all tasks
	tasks, err := mgr.ListTasks(sessionID, nil)
	if err != nil {
		return err
	}

	// Group by priority and status
	byPriority := make(map[string][]TaskExport)
	var doneTasks []TaskExport

	for _, task := range tasks {
		export := convertTaskToExport(task)
		if task.Status == "completed" {
			if includeDone {
				doneTasks = append(doneTasks, export)
			}
		} else {
			byPriority[task.Priority] = append(byPriority[task.Priority], export)
		}
	}

	// Write priority files
	totalCount := 0
	for _, priority := range []string{"p0", "p1", "p2", "p3", "p4"} {
		if tasks, ok := byPriority[priority]; ok && len(tasks) > 0 {
			filename := filepath.Join(dirPath, fmt.Sprintf("todo.%s.yaml", priority))
			if err := writeTasksToFile(tasks, filename); err != nil {
				return err
			}
			fmt.Printf("  todo.%s.yaml (%d tasks)\n", priority, len(tasks))
			totalCount += len(tasks)
		}
	}

	// Write done.yaml if requested
	if includeDone && len(doneTasks) > 0 {
		filename := filepath.Join(dirPath, "done.yaml")
		if err := writeTasksToFile(doneTasks, filename); err != nil {
			return err
		}
		fmt.Printf("  done.yaml (%d tasks)\n", len(doneTasks))
		totalCount += len(doneTasks)
	}

	fmt.Printf("\nExported %d tasks to %s/\n", totalCount, dirPath)
	fmt.Printf("\nImport back: llmtodo import --dir %s/\n", dirPath)
	return nil
}

// GenerateScaffold creates an empty directory structure with examples
func GenerateScaffold(dirPath string) error {
	// Create directory
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create todo.index.yaml
	indexContent := `# llmtodo file-based task tracking
# Compatible import: llmtodo import --dir todo/

system:
  description: "File-based todo tracking scaffold"
  import_command: "llmtodo import --dir todo/"
  export_command: "llmtodo export --dir todo/"

structure:
  todo.p0.yaml: "High priority tasks"
  todo.p1.yaml: "Important tasks"
  todo.p2.yaml: "Medium priority"
  todo.p3.yaml: "Low priority"
  todo.p4.yaml: "Optional tasks"
  done.yaml: "Completed tasks"
  deprecated.yaml: "Cancelled/archived tasks"

format:
  required_fields: ["title"]
  common_fields: ["priority", "effort", "files", "instructions"]
  see_examples: "Check todo.p0.yaml for working format"

workflow:
  - "Edit YAML files with your tasks"
  - "Run: llmtodo import --dir todo/"
  - "Tasks imported into llmtodo database"
  - "Work in database: llmtodo next, llmtodo done, etc."
  - "Export back: llmtodo export --dir todo/"
`
	if err := os.WriteFile(filepath.Join(dirPath, "todo.index.yaml"), []byte(indexContent), 0644); err != nil {
		return err
	}

	// Create example tasks
	exampleP0 := []TaskExport{
		{
			Title:    "Example high-priority task",
			Priority: "p0",
			Effort:   "s",
			Files:    []string{"path/to/file.go"},
			Instructions: map[string][]string{
				"must_do":     {"Step 1", "Step 2"},
				"must_not_do": {"Don't skip X"},
			},
		},
	}

	exampleP1 := []TaskExport{
		{
			Title:    "Example normal priority task",
			Priority: "p1",
			Effort:   "m",
		},
	}

	exampleDone := []TaskExport{
		{
			Title:    "Example completed task",
			Priority: "p0",
			Status:   "done",
		},
	}

	// Write example files
	writeTasksToFile(exampleP0, filepath.Join(dirPath, "todo.p0.yaml"))
	writeTasksToFile(exampleP1, filepath.Join(dirPath, "todo.p1.yaml"))
	writeTasksToFile(exampleDone, filepath.Join(dirPath, "done.yaml"))

	// Create empty files
	for _, p := range []string{"p2", "p3", "p4"} {
		os.WriteFile(filepath.Join(dirPath, fmt.Sprintf("todo.%s.yaml", p)), []byte("# Empty - add tasks here\n"), 0644)
	}
	os.WriteFile(filepath.Join(dirPath, "deprecated.yaml"), []byte("# Cancelled/archived tasks\n"), 0644)

	// Create README
	readmeContent := "# llm-todo File-Based Tracking\n\n" +
		"This directory contains file-based task tracking compatible with llmtodo.\n\n" +
		"## Structure\n\n" +
		"- `todo.p0.yaml` - High priority tasks\n" +
		"- `todo.p1.yaml` - Important tasks\n" +
		"- `todo.p2.yaml` - Medium priority\n" +
		"- `todo.p3.yaml` - Low priority\n" +
		"- `todo.p4.yaml` - Optional tasks\n" +
		"- `done.yaml` - Completed tasks\n" +
		"- `deprecated.yaml` - Cancelled/archived\n\n" +
		"## Workflow\n\n" +
		"1. Edit YAML files with your tasks\n" +
		"2. Import: `llmtodo import --dir todo/`\n" +
		"3. Work in database: `llmtodo next`, `llmtodo done`, etc.\n" +
		"4. Export back: `llmtodo export --dir todo/`\n\n" +
		"## Format\n\n" +
		"See `todo.p0.yaml` for working examples. Required field: `title`\n\n" +
		"Common fields:\n" +
		"- `priority`: p0, p1, p2, p3, p4\n" +
		"- `effort`: xs, s, m\n" +
		"- `files`: [\"path/to/file.go\"]\n" +
		"- `instructions.must_do`: [\"Step 1\", \"Step 2\"]\n"
	os.WriteFile(filepath.Join(dirPath, "README.md"), []byte(readmeContent), 0644)

	fmt.Printf("Created scaffold at %s/:\n", dirPath)
	fmt.Println("  todo.index.yaml (manifest)")
	fmt.Println("  todo.p0.yaml (1 example)")
	fmt.Println("  todo.p1.yaml (1 example)")
	fmt.Println("  todo.p2.yaml (empty)")
	fmt.Println("  todo.p3.yaml (empty)")
	fmt.Println("  todo.p4.yaml (empty)")
	fmt.Println("  done.yaml (1 example)")
	fmt.Println("  deprecated.yaml (empty)")
	fmt.Println("  README.md (guide)")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit YAML files in %s/\n", dirPath)
	fmt.Printf("  2. Import: llmtodo import --dir %s/\n", dirPath)

	return nil
}

// Helper functions

func convertTaskToExport(task *todo.Task) TaskExport {
	export := TaskExport{
		Title:    task.Task,
		Type:     task.Type,
		Priority: task.Priority,
		Status:   task.Status,
		Effort:   task.Effort,
	}

	// Parse JSON fields
	if task.Files != "" && task.Files != "[]" {
		json.Unmarshal([]byte(task.Files), &export.Files)
	}
	if task.Instructions != "" && task.Instructions != "{}" {
		json.Unmarshal([]byte(task.Instructions), &export.Instructions)
	}
	if task.Refs != "" && task.Refs != "[]" {
		json.Unmarshal([]byte(task.Refs), &export.Refs)
	}

	return export
}

func writeTasksToFile(tasks []TaskExport, filename string) error {
	data, err := yaml.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
