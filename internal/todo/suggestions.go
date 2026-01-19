package todo

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Suggestion represents a smart suggestion for the user
type Suggestion struct {
	Type    string // "modified_files", "duplicate", "prerequisite"
	Message string
	TaskID  int
}

// GetSuggestions returns smart suggestions based on current context
func (m *Manager) GetSuggestions(sessionID string) ([]Suggestion, error) {
	var suggestions []Suggestion

	// Check for modified files that match task files
	modifiedFiles, _ := getModifiedFiles()
	if len(modifiedFiles) > 0 {
		fileSuggestions, _ := m.findTasksByFiles(sessionID, modifiedFiles)
		suggestions = append(suggestions, fileSuggestions...)
	}

	return suggestions, nil
}

func getModifiedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

func (m *Manager) findTasksByFiles(sessionID string, modifiedFiles []string) ([]Suggestion, error) {
	tasks, err := m.ListTasks(sessionID, map[string]string{})
	if err != nil {
		return nil, err
	}

	var suggestions []Suggestion

	for _, task := range tasks {
		if task.Files == "" || task.Files == "[]" {
			continue
		}

		var taskFiles []string
		if err := json.Unmarshal([]byte(task.Files), &taskFiles); err != nil {
			continue
		}

		// Check if any modified file matches task files
		for _, modFile := range modifiedFiles {
			for _, taskFile := range taskFiles {
				if strings.Contains(modFile, taskFile) || strings.Contains(taskFile, modFile) {
					msg := fmt.Sprintf("Modified %s - related to task #%d (%s)", modFile, task.ID, task.Task)
					suggestions = append(suggestions, Suggestion{
						Type:    "modified_files",
						Message: msg,
						TaskID:  task.ID,
					})
					break
				}
			}
		}
	}

	return suggestions, nil
}
