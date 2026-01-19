package todo

import (
	"strings"
)

// SearchTasks searches tasks by keyword across multiple fields
func (m *Manager) SearchTasks(sessionID, query string, includeCompleted bool) ([]*Task, error) {
	// Default: search all statuses (including completed)
	filters := make(map[string]string)
	if !includeCompleted {
		// Exclude completed if explicitly requested
		// (but by default we want to search completed)
		filters["status"] = "pending"
	}

	tasks, err := m.ListTasks(sessionID, filters)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var matches []*Task

	for _, task := range tasks {
		if matchesQuery(task, queryLower) {
			matches = append(matches, task)
		}
	}

	return matches, nil
}

func matchesQuery(task *Task, query string) bool {
	// Search in task title
	if strings.Contains(strings.ToLower(task.Task), query) {
		return true
	}

	// Search in active form
	if strings.Contains(strings.ToLower(task.ActiveForm), query) {
		return true
	}

	// Search in notes
	if strings.Contains(strings.ToLower(task.Notes), query) {
		return true
	}

	// Search in files
	if strings.Contains(strings.ToLower(task.Files), query) {
		return true
	}

	// Search in refs
	if strings.Contains(strings.ToLower(task.Refs), query) {
		return true
	}

	// Search in instructions
	if strings.Contains(strings.ToLower(task.Instructions), query) {
		return true
	}

	return false
}
