package todo

import (
	"fmt"
	"os"
	"path/filepath"
)

// SessionSummary represents a session with stats
type SessionSummary struct {
	ID              string
	Type            string
	Goal            string
	Total           int
	Pending         int
	InProgress      int
	Completed       int
	Blocked         int
	UpdatedAt       string
}

// ListSessions returns all sessions with stats
func (m *Manager) ListSessions() ([]SessionSummary, error) {
	rows, err := m.db.Query(`
		SELECT
			s.id,
			s.type,
			COALESCE(s.goal, ''),
			s.updated_at,
			COALESCE(SUM(CASE WHEN t.status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			COALESCE(SUM(CASE WHEN t.status = 'in_progress' THEN 1 ELSE 0 END), 0) as in_progress,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN t.status = 'blocked' THEN 1 ELSE 0 END), 0) as blocked,
			COUNT(t.id) as total
		FROM sessions s
		LEFT JOIN todos t ON s.id = t.session_id
		GROUP BY s.id
		ORDER BY s.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		var s SessionSummary
		if err := rows.Scan(&s.ID, &s.Type, &s.Goal, &s.UpdatedAt, &s.Pending, &s.InProgress, &s.Completed, &s.Blocked, &s.Total); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// GetCurrentSession returns the active session ID from ~/.llm-todo/current
func GetCurrentSession() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	currentFile := filepath.Join(home, ".llm-todo", "current")
	data, err := os.ReadFile(currentFile)
	if err != nil {
		return "", fmt.Errorf("no active session (use 'todo use <session>' or run from project directory)")
	}

	return string(data), nil
}

// SetCurrentSession sets the active session
func SetCurrentSession(sessionID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	currentFile := filepath.Join(home, ".llm-todo", "current")
	dir := filepath.Dir(currentFile)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(currentFile, []byte(sessionID), 0644)
}
