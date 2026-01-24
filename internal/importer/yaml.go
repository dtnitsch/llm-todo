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
	ID           string              `yaml:"id"`
	Title        string              `yaml:"title"`
	Type         string              `yaml:"type,omitempty"`
	Priority     string              `yaml:"priority,omitempty"`
	Status       string              `yaml:"status,omitempty"`
	Effort       string              `yaml:"effort,omitempty"`
	Description  string              `yaml:"description,omitempty"`
	DependsOn    []string            `yaml:"depends_on,omitempty"`
	Instructions map[string][]string `yaml:"instructions,omitempty"`
	Files        []string            `yaml:"files,omitempty"`
	Refs         []string            `yaml:"refs,omitempty"`
}

// YAMLFile represents the top-level YAML file structure
type YAMLFile struct {
	Goal  string     `yaml:"goal,omitempty"`
	Tasks []YAMLTask `yaml:"tasks,omitempty"`
}

// UpdateTasksFromYAML updates existing tasks from a YAML file by ID and returns (count, goal, error)
func UpdateTasksFromYAML(mgr *todo.Manager, sessionID, filePath string) (int, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("❌ Failed to read file: %s\n   Make sure the file exists and is readable", filePath)
	}

	// Check for common typos before parsing
	typos := CheckForCommonTypos(string(data))
	if len(typos) > 0 {
		return 0, "", fmt.Errorf("❌ Possible typos detected:\n   %s\n\n   Fix these and try again", strings.Join(typos, "\n   "))
	}

	// Parse YAML file
	var yamlFile YAMLFile
	var tasks []YAMLTask
	var goal string

	if err := yaml.Unmarshal(data, &yamlFile); err == nil && len(yamlFile.Tasks) > 0 {
		tasks = yamlFile.Tasks
		goal = yamlFile.Goal
	} else {
		if err := yaml.Unmarshal(data, &tasks); err != nil {
			return 0, "", fmt.Errorf("❌ Invalid YAML format:\n   %v\n\n   💡 Check indentation (use spaces, not tabs)\n   💡 Run: llmtodo import --template to see correct format", err)
		}
	}

	count := 0
	var errors []string
	var skippedTasks []string

	for _, yamlTask := range tasks {
		if yamlTask.ID == "" {
			skippedTasks = append(skippedTasks, fmt.Sprintf("Skipped task without ID: %q", yamlTask.Title))
			continue
		}

		// Extract task number from "task-{num}"
		var taskID int
		if _, err := fmt.Sscanf(yamlTask.ID, "task-%d", &taskID); err != nil {
			skippedTasks = append(skippedTasks, fmt.Sprintf("Skipped malformed ID: %q (expected format: task-123)", yamlTask.ID))
			continue
		}

		// Validate fields before updating
		if yamlTask.Priority != "" {
			if err := ValidatePriority(yamlTask.Priority); err != nil {
				ie := err.(*ImportError)
				ie.TaskID = yamlTask.ID
				errors = append(errors, ie.Error())
				continue
			}
		}

		if yamlTask.Effort != "" {
			if err := ValidateEffort(yamlTask.Effort); err != nil {
				ie := err.(*ImportError)
				ie.TaskID = yamlTask.ID
				errors = append(errors, ie.Error())
				continue
			}
		}

		if yamlTask.Type != "" {
			if err := ValidateTaskType(yamlTask.Type); err != nil {
				ie := err.(*ImportError)
				ie.TaskID = yamlTask.ID
				errors = append(errors, ie.Error())
				continue
			}
		}

		// Build updates map (only non-empty fields)
		updates := make(map[string]interface{})

		if yamlTask.Title != "" {
			updates["task"] = yamlTask.Title
			updates["active_form"] = generateActiveForm(yamlTask.Title)
		}

		if yamlTask.Priority != "" {
			updates["priority"] = yamlTask.Priority
		}

		if yamlTask.Effort != "" {
			updates["effort"] = yamlTask.Effort
		}

		if yamlTask.Status != "" {
			updates["status"] = mapStatus(yamlTask.Status)
		}

		if yamlTask.Type != "" {
			updates["type"] = yamlTask.Type
		}

		// Only update files if non-empty array
		if len(yamlTask.Files) > 0 {
			filesJSON, _ := json.Marshal(yamlTask.Files)
			updates["files"] = string(filesJSON)
		}

		// Only update refs if non-empty array
		if len(yamlTask.Refs) > 0 {
			refsJSON, _ := json.Marshal(yamlTask.Refs)
			updates["refs"] = string(refsJSON)
		}

		// Only update instructions if non-empty
		if len(yamlTask.Instructions) > 0 {
			instructionsJSON, _ := json.Marshal(yamlTask.Instructions)
			updates["instructions"] = string(instructionsJSON)
		}

		// Only update if we have changes
		if len(updates) > 0 {
			if err := mgr.UpdateTask(taskID, updates); err != nil {
				// Task doesn't exist
				errors = append(errors, fmt.Sprintf("❌ Task %s not found\n   Task ID %d doesn't exist in this session\n   💡 Run: llmtodo get pending to see available tasks", yamlTask.ID, taskID))
				continue
			}
			count++
		}
	}

	// If we had errors, return them
	if len(errors) > 0 {
		return count, goal, fmt.Errorf("\n%s", strings.Join(errors, "\n\n"))
	}

	// Show warnings for skipped tasks (but don't fail)
	if len(skippedTasks) > 0 && count == 0 {
		return 0, goal, fmt.Errorf("⚠️  No tasks updated. Issues:\n   %s", strings.Join(skippedTasks, "\n   "))
	}

	return count, goal, nil
}

// ImportFromYAML imports tasks from a YAML file and returns (count, goal, error)
func ImportFromYAML(mgr *todo.Manager, sessionID, filePath string) (int, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("❌ Failed to read file: %s\n   Make sure the file exists and is readable", filePath)
	}

	// Check for common typos before parsing
	typos := CheckForCommonTypos(string(data))
	if len(typos) > 0 {
		return 0, "", fmt.Errorf("❌ Possible typos detected:\n   %s\n\n   Fix these and try again", strings.Join(typos, "\n   "))
	}

	// Try parsing as YAMLFile first (new format with goal)
	var yamlFile YAMLFile
	var tasks []YAMLTask
	var goal string

	if err := yaml.Unmarshal(data, &yamlFile); err == nil && len(yamlFile.Tasks) > 0 {
		// New format with goal
		tasks = yamlFile.Tasks
		goal = yamlFile.Goal
	} else {
		// Fallback: try parsing as array of tasks (old format)
		if err := yaml.Unmarshal(data, &tasks); err != nil {
			return 0, "", fmt.Errorf("❌ Invalid YAML format:\n   %v\n\n   💡 Check indentation (use spaces, not tabs)\n   💡 Run: llmtodo import --template to see correct format", err)
		}
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
			return count, goal, fmt.Errorf("failed to create task: %w", err)
		}

		if yamlTask.ID != "" {
			idMap[yamlTask.ID] = id
		}

		count++
	}

	return count, goal, nil
}

// ImportFromDirectory imports all YAML files from a directory
func ImportFromDirectory(mgr *todo.Manager, sessionID, dirPath string) (int, string, error) {
	files := []string{"todo.p0.yaml", "todo.p1.yaml", "todo.p2.yaml", "todo.p3.yaml", "todo.p4.yaml"}

	totalCount := 0
	var firstGoal string
	for _, file := range files {
		filePath := filepath.Join(dirPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		count, goal, err := ImportFromYAML(mgr, sessionID, filePath)
		if err != nil {
			return totalCount, firstGoal, err
		}
		totalCount += count

		// Keep the first non-empty goal we find
		if firstGoal == "" && goal != "" {
			firstGoal = goal
		}
	}

	return totalCount, firstGoal, nil
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
