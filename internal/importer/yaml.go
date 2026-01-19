package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtnitsch/llm-todo/internal/todo"
	"gopkg.in/yaml.v3"
)

// YAMLTask represents a task in YAML format
type YAMLTask struct {
	ID           string            `yaml:"id"`
	Title        string            `yaml:"title"`
	Type         string            `yaml:"type,omitempty"`
	Priority     string            `yaml:"priority,omitempty"`
	Status       string            `yaml:"status,omitempty"`
	Effort       string            `yaml:"effort,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	DependsOn    []string          `yaml:"depends_on,omitempty"`
	Instructions map[string][]string `yaml:"instructions,omitempty"`
	Files        []string          `yaml:"files,omitempty"`
	Refs         []string          `yaml:"refs,omitempty"`
}

// ImportFromYAML imports tasks from a YAML file
func ImportFromYAML(mgr *todo.Manager, sessionID, filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	var tasks []YAMLTask
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return 0, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract priority from filename (e.g., todo.p0.yaml -> p0)
	filePriority := extractPriorityFromFilename(filePath)

	// Map logical IDs to database IDs
	idMap := make(map[string]int64)

	count := 0
	for _, yamlTask := range tasks {
		if yamlTask.Title == "" {
			continue
		}

		// Use file priority if task doesn't specify one
		if yamlTask.Priority == "" {
			yamlTask.Priority = filePriority
		}

		// Convert to Task
		task := convertYAMLTask(yamlTask, sessionID, idMap)

		id, err := mgr.CreateTask(task)
		if err != nil {
			return count, fmt.Errorf("failed to create task: %w", err)
		}

		if yamlTask.ID != "" {
			idMap[yamlTask.ID] = id
		}

		count++
	}

	return count, nil
}

// ImportFromDirectory imports all YAML files from a directory
func ImportFromDirectory(mgr *todo.Manager, sessionID, dirPath string) (int, error) {
	files := []string{"todo.p0.yaml", "todo.p1.yaml", "todo.p2.yaml", "todo.p3.yaml", "todo.p4.yaml"}

	totalCount := 0
	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		count, err := ImportFromYAML(mgr, sessionID, filePath)
		if err != nil {
			return totalCount, err
		}
		totalCount += count
	}

	return totalCount, nil
}

func convertYAMLTask(yamlTask YAMLTask, sessionID string, idMap map[string]int64) *todo.Task {
	// Map status
	status := mapStatus(yamlTask.Status)

	// Default priority
	priority := yamlTask.Priority
	if priority == "" {
		priority = "p1"
	}

	// Default type
	taskType := yamlTask.Type
	if taskType == "" {
		taskType = "task"
	}

	// Convert dependencies
	dependantIDs := "[]"
	if len(yamlTask.DependsOn) > 0 {
		var dbIDs []int64
		for _, logicalID := range yamlTask.DependsOn {
			if dbID, ok := idMap[logicalID]; ok {
				dbIDs = append(dbIDs, dbID)
			}
		}
		if len(dbIDs) > 0 {
			idsJSON, _ := json.Marshal(dbIDs)
			dependantIDs = string(idsJSON)
		}
	}

	// Convert instructions
	instructions := ""
	if len(yamlTask.Instructions) > 0 {
		instructionsJSON, _ := json.Marshal(yamlTask.Instructions)
		instructions = string(instructionsJSON)
	}

	// Convert files
	files := ""
	if len(yamlTask.Files) > 0 {
		filesJSON, _ := json.Marshal(yamlTask.Files)
		files = string(filesJSON)
	}

	// Convert refs
	refs := ""
	if len(yamlTask.Refs) > 0 {
		refsJSON, _ := json.Marshal(yamlTask.Refs)
		refs = string(refsJSON)
	}

	return &todo.Task{
		SessionID:    sessionID,
		Type:         taskType,
		Priority:     priority,
		Status:       status,
		Task:         yamlTask.Title,
		ActiveForm:   generateActiveForm(yamlTask.Title),
		Files:        files,
		Refs:         refs,
		Instructions: instructions,
		DependantIDs: dependantIDs,
		Effort:       yamlTask.Effort,
		Metadata:     "{}",
	}
}

func mapStatus(yamlStatus string) string {
	switch yamlStatus {
	case "pending", "todo":
		return "pending"
	case "in_progress", "in-progress":
		return "in_progress"
	case "completed", "done":
		return "completed"
	case "blocked":
		return "blocked"
	default:
		return "pending"
	}
}

func generateActiveForm(title string) string {
	verbs := map[string]string{
		"Create":    "Creating",
		"Implement": "Implementing",
		"Build":     "Building",
		"Fix":       "Fixing",
		"Update":    "Updating",
		"Add":       "Adding",
		"Remove":    "Removing",
		"Refactor":  "Refactoring",
		"Extract":   "Extracting",
	}

	for verb, ing := range verbs {
		if strings.HasPrefix(title, verb+" ") {
			return ing + strings.TrimPrefix(title, verb)
		}
	}

	return "Working on: " + title
}

func extractPriorityFromFilename(filePath string) string {
	filename := filepath.Base(filePath)

	// Match patterns like: todo.p0.yaml, p1.yaml, etc.
	if strings.Contains(filename, ".p0.") || strings.HasSuffix(filename, "p0.yaml") {
		return "p0"
	}
	if strings.Contains(filename, ".p1.") || strings.HasSuffix(filename, "p1.yaml") {
		return "p1"
	}
	if strings.Contains(filename, ".p2.") || strings.HasSuffix(filename, "p2.yaml") {
		return "p2"
	}
	if strings.Contains(filename, ".p3.") || strings.HasSuffix(filename, "p3.yaml") {
		return "p3"
	}
	if strings.Contains(filename, ".p4.") || strings.HasSuffix(filename, "p4.yaml") {
		return "p4"
	}

	return "p1" // Default
}
