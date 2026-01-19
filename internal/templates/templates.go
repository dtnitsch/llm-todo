package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

//go:embed embedded/*.yaml
var embeddedFS embed.FS

// LoadTemplate loads a template by name from embedded or user templates
func LoadTemplate(name string) (*Template, error) {
	// Try user templates first (~/.llm-todo/templates/)
	userPath := getUserTemplatePath(name)
	if _, err := os.Stat(userPath); err == nil {
		return loadFromFile(userPath)
	}

	// Try embedded templates
	embeddedPath := fmt.Sprintf("embedded/%s.yaml", name)
	data, err := embeddedFS.ReadFile(embeddedPath)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &tmpl, nil
}

// ListTemplates returns all available templates (embedded + user)
func ListTemplates() ([]TemplateInfo, error) {
	var templates []TemplateInfo

	// List embedded templates
	entries, err := embeddedFS.ReadDir("embedded")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
				name := strings.TrimSuffix(entry.Name(), ".yaml")
				tmpl, err := LoadTemplate(name)
				if err == nil {
					templates = append(templates, TemplateInfo{
						Name:        name,
						Description: tmpl.Description,
						TaskCount:   len(tmpl.Tasks),
						Source:      "embedded",
					})
				}
			}
		}
	}

	// List user templates
	userDir := getUserTemplateDir()
	if entries, err := os.ReadDir(userDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
				name := strings.TrimSuffix(entry.Name(), ".yaml")
				tmpl, err := LoadTemplate(name)
				if err == nil {
					templates = append(templates, TemplateInfo{
						Name:        name,
						Description: tmpl.Description,
						TaskCount:   len(tmpl.Tasks),
						Source:      "user",
					})
				}
			}
		}
	}

	return templates, nil
}

// ApplyTemplate creates tasks from a template in the specified session
func ApplyTemplate(mgr *todo.Manager, sessionID string, templateName string, vars map[string]string) (int, error) {
	tmpl, err := LoadTemplate(templateName)
	if err != nil {
		return 0, err
	}

	// Track logical IDs to database IDs for dependency resolution
	idMap := make(map[string]int64)
	created := 0

	for i, taskSpec := range tmpl.Tasks {
		task := &todo.Task{
			SessionID: sessionID,
			Task:      substituteVars(taskSpec.Title, vars),
			Priority:  taskSpec.Priority,
			Effort:    taskSpec.Effort,
			Type:      taskSpec.Type,
			Status:    "pending",
		}

		// Set default priority if not specified
		if task.Priority == "" {
			task.Priority = "p0"
		}

		// Convert files array to JSON
		if len(taskSpec.Files) > 0 {
			files := make([]string, len(taskSpec.Files))
			for j, f := range taskSpec.Files {
				files[j] = substituteVars(f, vars)
			}
			task.Files = marshalJSON(files)
		}

		// Convert instructions to JSON
		if len(taskSpec.Instructions) > 0 {
			instructions := make(map[string][]string)
			for key, values := range taskSpec.Instructions {
				substituted := make([]string, len(values))
				for j, v := range values {
					substituted[j] = substituteVars(v, vars)
				}
				instructions[key] = substituted
			}
			task.Instructions = marshalJSON(instructions)
		}

		// Create task
		id, err := mgr.CreateTask(task)
		if err != nil {
			return created, fmt.Errorf("failed to create task %d: %w", i+1, err)
		}

		// Store logical ID mapping for dependency resolution
		logicalID := fmt.Sprintf("task-%d", i)
		idMap[logicalID] = id
		created++

		// Update dependencies if this task has depends_on
		if len(taskSpec.DependsOn) > 0 {
			var depIDs []int64
			for _, depName := range taskSpec.DependsOn {
				if depID, exists := idMap[depName]; exists {
					depIDs = append(depIDs, depID)
				}
			}
			if len(depIDs) > 0 {
				mgr.UpdateTask(int(id), map[string]interface{}{
					"dependant_ids": marshalJSON(depIDs),
				})
			}
		}
	}

	return created, nil
}

// Helper functions

func getUserTemplateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".llm-todo", "templates")
}

func getUserTemplatePath(name string) string {
	return filepath.Join(getUserTemplateDir(), name+".yaml")
}

func loadFromFile(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &tmpl, nil
}

func substituteVars(s string, vars map[string]string) string {
	result := s
	for key, value := range vars {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func marshalJSON(v interface{}) string {
	// Simple JSON marshaling (reuse from existing code)
	// This is a simplified version - in production, use json.Marshal
	switch val := v.(type) {
	case []string:
		if len(val) == 0 {
			return "[]"
		}
		var parts []string
		for _, s := range val {
			parts = append(parts, fmt.Sprintf(`"%s"`, s))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case []int64:
		if len(val) == 0 {
			return "[]"
		}
		var parts []string
		for _, n := range val {
			parts = append(parts, fmt.Sprintf("%d", n))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case map[string][]string:
		if len(val) == 0 {
			return "{}"
		}
		var parts []string
		for key, values := range val {
			valuesJSON := marshalJSON(values)
			parts = append(parts, fmt.Sprintf(`"%s":%s`, key, valuesJSON))
		}
		return "{" + strings.Join(parts, ",") + "}"
	}
	return "{}"
}
