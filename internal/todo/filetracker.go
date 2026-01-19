package todo

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// TrackModifiedFiles captures files modified in git and stores them in task
func (m *Manager) TrackModifiedFiles(taskID int) ([]string, error) {
	// Get modified files from git
	modifiedFiles, err := getGitModifiedFiles()
	if err != nil || len(modifiedFiles) == 0 {
		// Not in git repo or no modifications - not an error
		return nil, nil
	}

	// Get current task
	task, err := m.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// Merge with existing files
	var existingFiles []string
	if task.Files != "" && task.Files != "[]" {
		json.Unmarshal([]byte(task.Files), &existingFiles)
	}

	// Combine and deduplicate
	fileSet := make(map[string]bool)
	for _, f := range existingFiles {
		fileSet[f] = true
	}
	for _, f := range modifiedFiles {
		fileSet[f] = true
	}

	var allFiles []string
	for f := range fileSet {
		allFiles = append(allFiles, f)
	}

	// Update task with all files
	filesJSON, _ := json.Marshal(allFiles)
	if err := m.UpdateTask(taskID, map[string]interface{}{"files": string(filesJSON)}); err != nil {
		return nil, err
	}

	return modifiedFiles, nil
}

func getGitModifiedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
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
