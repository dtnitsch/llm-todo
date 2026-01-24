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

// ListSessions returns sessions with stats, optionally filtered by archived status
// includeArchived: if true, returns all sessions; if false, excludes archived sessions
func (m *Manager) ListSessions(includeArchived bool) ([]SessionSummary, error) {
	query := `
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
		LEFT JOIN todos t ON s.id = t.session_id`

	if !includeArchived {
		query += `
		WHERE s.status != 'archived'`
	}

	query += `
		GROUP BY s.id
		ORDER BY s.updated_at DESC
	`

	rows, err := m.db.Query(query)
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

// ArchiveSession archives a session (sets status to 'archived')
func (m *Manager) ArchiveSession(sessionID string) error {
	updates := map[string]interface{}{
		"status": "archived",
	}
	return m.UpdateSession(sessionID, updates)
}

// RestoreSession restores an archived session (sets status to 'active')
func (m *Manager) RestoreSession(sessionID string) error {
	updates := map[string]interface{}{
		"status": "active",
	}
	return m.UpdateSession(sessionID, updates)
}

// GetCurrentSession returns the active session ID from project-local or global current file
func GetCurrentSession() (string, error) {
	currentFile := getCurrentFilePath()
	data, err := os.ReadFile(currentFile)
	if err != nil {
		return "", fmt.Errorf("no active session (use 'todo use <session>' or run from project directory)")
	}

	return string(data), nil
}

// SetCurrentSession sets the active session
func SetCurrentSession(sessionID string) error {
	currentFile := getCurrentFilePath()
	dir := filepath.Dir(currentFile)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(currentFile, []byte(sessionID), 0644)
}

// getCurrentFilePath returns the path to the current session file (project-local or global)
func getCurrentFilePath() string {
	// Check if we're in a project directory
	if isProjectDir() {
		return ".llm-todo/current"
	}

	// Fall back to global
	home, err := os.UserHomeDir()
	if err != nil {
		return ".llm-todo/current"
	}

	return filepath.Join(home, ".llm-todo", "current")
}

// isProjectDir checks if the current directory is a project directory
func isProjectDir() bool {
	// Check for .git directory (git repository)
	if _, err := os.Stat(".git"); err == nil {
		return true
	}
	// Check for .llm-todo directory (explicit opt-in)
	if _, err := os.Stat(".llm-todo"); err == nil {
		return true
	}
	return false
}
