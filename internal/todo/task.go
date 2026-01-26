package todo

import (
	"database/sql"
	"fmt"
)

// CreateTask creates a new task
func (m *Manager) CreateTask(task *Task) (int64, error) {
	// Auto-assign priority_order if not set
	if task.PriorityOrder == 0 {
		maxOrder, err := m.getMaxPriorityOrder(task.SessionID, task.Priority)
		if err != nil {
			return 0, err
		}
		task.PriorityOrder = maxOrder + 100
	}

	result, err := m.db.Exec(`
		INSERT INTO todos (
			session_id, type, priority, priority_order, status, task, active_form,
			files, refs, waiting_on, output, audience,
			instructions, notes, blocking_reason, dependant_ids, effort, metadata
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.SessionID, task.Type, task.Priority, task.PriorityOrder, task.Status, task.Task, task.ActiveForm,
		task.Files, task.Refs, task.WaitingOn, task.Output, task.Audience,
		task.Instructions, task.Notes, task.BlockingReason, task.DependantIDs, task.Effort, task.Metadata)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// BulkCreateTasks creates multiple tasks in a single atomic INSERT
// Returns slice of task IDs in the same order as input tasks
func (m *Manager) BulkCreateTasks(tasks []*Task) ([]int64, error) {
	if len(tasks) == 0 {
		return []int64{}, nil
	}

	// Get max priority orders for all (sessionID, priority) combinations
	priorityOrders := make(map[string]int)
	for _, task := range tasks {
		key := task.SessionID + ":" + task.Priority
		if _, exists := priorityOrders[key]; !exists {
			maxOrder, err := m.getMaxPriorityOrder(task.SessionID, task.Priority)
			if err != nil {
				return nil, err
			}
			priorityOrders[key] = maxOrder
		}
	}

	// Assign priority_order to each task
	orderCounters := make(map[string]int)
	for _, task := range tasks {
		if task.PriorityOrder == 0 {
			key := task.SessionID + ":" + task.Priority
			baseOrder := priorityOrders[key]
			orderCounters[key]++
			task.PriorityOrder = baseOrder + (orderCounters[key] * 100)
		}
	}

	// Build multi-row INSERT statement
	query := `INSERT INTO todos (
		session_id, type, priority, priority_order, status, task, active_form,
		files, refs, waiting_on, output, audience,
		instructions, notes, blocking_reason, dependant_ids, effort, metadata
	) VALUES `

	valuePlaceholders := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	args := []interface{}{}

	for i, task := range tasks {
		if i > 0 {
			query += ", "
		}
		query += valuePlaceholders

		args = append(args,
			task.SessionID, task.Type, task.Priority, task.PriorityOrder, task.Status, task.Task, task.ActiveForm,
			task.Files, task.Refs, task.WaitingOn, task.Output, task.Audience,
			task.Instructions, task.Notes, task.BlockingReason, task.DependantIDs, task.Effort, task.Metadata)
	}

	// Execute bulk insert
	result, err := m.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("bulk insert failed: %w", err)
	}

	// Get the last inserted ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Calculate all IDs
	// Note: SQLite's LastInsertId() returns the LAST row inserted in a multi-row INSERT
	// So we calculate backward from lastID to get all IDs: [lastID-N+1, lastID-N+2, ..., lastID]
	ids := make([]int64, len(tasks))
	for i := range tasks {
		ids[i] = lastID - int64(len(tasks)-1-i)
	}

	return ids, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID int) (*Task, error) {
	row := m.db.QueryRow(`
		SELECT id, session_id, type, priority, priority_order, status, task, active_form,
		       COALESCE(files, ''), COALESCE(refs, ''), COALESCE(waiting_on, ''),
		       COALESCE(output, ''), COALESCE(audience, ''),
		       COALESCE(instructions, ''), COALESCE(notes, ''), COALESCE(blocking_reason, ''),
		       COALESCE(dependant_ids, ''), COALESCE(effort, ''), COALESCE(metadata, '{}'),
		       created_at, updated_at
		FROM todos
		WHERE id = ?
	`, taskID)

	var t Task
	err := row.Scan(&t.ID, &t.SessionID, &t.Type, &t.Priority, &t.PriorityOrder, &t.Status, &t.Task, &t.ActiveForm,
		&t.Files, &t.Refs, &t.WaitingOn, &t.Output, &t.Audience,
		&t.Instructions, &t.Notes, &t.BlockingReason, &t.DependantIDs, &t.Effort, &t.Metadata,
		&t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %d", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &t, nil
}

// UpdateTask updates task fields
func (m *Manager) UpdateTask(taskID int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE todos SET "
	args := []interface{}{}

	first := true
	for field, value := range updates {
		if !first {
			query += ", "
		}
		query += field + " = ?"
		args = append(args, value)
		first = false
	}

	query += ", updated_at = datetime('now') WHERE id = ?"
	args = append(args, taskID)

	result, err := m.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found: %d", taskID)
	}

	return nil
}

// UpdateTaskInSession updates a task and validates it belongs to the specified session
func (m *Manager) UpdateTaskInSession(sessionID string, taskID int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE todos SET "
	args := []interface{}{}

	first := true
	for field, value := range updates {
		if !first {
			query += ", "
		}
		query += field + " = ?"
		args = append(args, value)
		first = false
	}

	query += ", updated_at = datetime('now') WHERE id = ? AND session_id = ?"
	args = append(args, taskID, sessionID)

	result, err := m.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task %d not found in session %s", taskID, sessionID)
	}

	return nil
}

// UpdateTaskStatus updates a task's status
func (m *Manager) UpdateTaskStatus(taskID int, status string) error {
	return m.UpdateTask(taskID, map[string]interface{}{"status": status})
}

// DeleteTask permanently deletes a task from the database
func (m *Manager) DeleteTask(taskID int) error {
	result, err := m.db.Exec("DELETE FROM todos WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task not found: %d", taskID)
	}

	return nil
}

func (m *Manager) getMaxPriorityOrder(sessionID, priority string) (int, error) {
	var maxOrder sql.NullInt64
	err := m.db.QueryRow(`
		SELECT MAX(priority_order)
		FROM todos
		WHERE session_id = ? AND priority = ?
	`, sessionID, priority).Scan(&maxOrder)

	if err != nil {
		return 0, err
	}

	if !maxOrder.Valid {
		return 0, nil
	}

	return int(maxOrder.Int64), nil
}
