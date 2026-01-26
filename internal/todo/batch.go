package todo

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTaskIDs parses comma-separated task IDs
func ParseTaskIDs(idsStr string) ([]int, error) {
	parts := strings.Split(idsStr, ",")
	var ids []int

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid task ID: %s", part)
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no task IDs provided")
	}

	return ids, nil
}

// BatchUpdateStatus updates status for multiple tasks
func (m *Manager) BatchUpdateStatus(taskIDs []int, status string) error {
	for _, id := range taskIDs {
		if err := m.UpdateTaskStatus(id, status); err != nil {
			return fmt.Errorf("failed to update task %d: %w", id, err)
		}
	}
	return nil
}

// BatchUpdatePriority updates priority_order for multiple tasks
func (m *Manager) BatchUpdatePriority(taskIDs []int, priorityOrder int) error {
	for _, id := range taskIDs {
		if err := m.UpdateTask(id, map[string]interface{}{"priority_order": priorityOrder}); err != nil {
			return fmt.Errorf("failed to update task %d: %w", id, err)
		}
	}
	return nil
}

// BatchAddNote adds a note to multiple tasks
func (m *Manager) BatchAddNote(taskIDs []int, note string) error {
	for _, id := range taskIDs {
		task, err := m.GetTask(id)
		if err != nil {
			return fmt.Errorf("failed to get task %d: %w", id, err)
		}

		existingNotes := task.Notes
		if existingNotes != "" {
			existingNotes += "\n\n"
		}
		existingNotes += note

		if err := m.UpdateTask(id, map[string]interface{}{"notes": existingNotes}); err != nil {
			return fmt.Errorf("failed to update task %d: %w", id, err)
		}
	}
	return nil
}

// BatchDeleteTasks permanently deletes multiple tasks from the database
func (m *Manager) BatchDeleteTasks(taskIDs []int) error {
	for _, id := range taskIDs {
		if err := m.DeleteTask(id); err != nil {
			return fmt.Errorf("failed to delete task %d: %w", id, err)
		}
	}
	return nil
}
