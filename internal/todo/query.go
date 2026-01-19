package todo

import "fmt"

// ListTasks lists tasks for a session with optional filters
func (m *Manager) ListTasks(sessionID string, filters map[string]string) ([]*Task, error) {
	query := `
		SELECT id, session_id, type, priority, priority_order, status, task, active_form,
		       COALESCE(files, ''), COALESCE(refs, ''), COALESCE(waiting_on, ''),
		       COALESCE(output, ''), COALESCE(audience, ''),
		       COALESCE(instructions, ''), COALESCE(notes, ''), COALESCE(blocking_reason, ''),
		       COALESCE(dependant_ids, ''), COALESCE(effort, ''), COALESCE(metadata, '{}'),
		       created_at, updated_at
		FROM todos
		WHERE session_id = ?
	`
	args := []interface{}{sessionID}

	if priority, ok := filters["priority"]; ok && priority != "" {
		query += " AND priority = ?"
		args = append(args, priority)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if taskType, ok := filters["type"]; ok && taskType != "" {
		query += " AND type = ?"
		args = append(args, taskType)
	}

	query += " ORDER BY priority, priority_order, id"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.SessionID, &t.Type, &t.Priority, &t.PriorityOrder, &t.Status, &t.Task, &t.ActiveForm,
			&t.Files, &t.Refs, &t.WaitingOn, &t.Output, &t.Audience,
			&t.Instructions, &t.Notes, &t.BlockingReason, &t.DependantIDs, &t.Effort, &t.Metadata,
			&t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

// GetStats returns task statistics for a session
func (m *Manager) GetStats(sessionID string) (map[string]int, error) {
	stats := make(map[string]int)

	rows, err := m.db.Query(`
		SELECT status, COUNT(*)
		FROM todos
		WHERE session_id = ?
		GROUP BY status
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	var total int
	if err := m.db.QueryRow("SELECT COUNT(*) FROM todos WHERE session_id = ?", sessionID).Scan(&total); err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}

// GetNextTask returns the next pending task (first by priority order)
func (m *Manager) GetNextTask(sessionID string) (*Task, error) {
	tasks, err := m.ListTasks(sessionID, map[string]string{"status": "pending"})
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no pending tasks")
	}

	return tasks[0], nil
}
