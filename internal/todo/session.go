package todo

import (
	"database/sql"
	"fmt"
)

// CreateSession creates a new session
func (m *Manager) CreateSession(session *Session) error {
	_, err := m.db.Exec(`
		INSERT INTO sessions (id, type, goal, success_criteria, boundaries, deliverables, status, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, session.ID, session.Type, session.Goal, session.SuccessCriteria, session.Boundaries, session.Deliverables, session.Status, session.Metadata)

	return err
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	row := m.db.QueryRow(`
		SELECT id, type, goal, success_criteria, boundaries, deliverables, status, metadata, created_at, updated_at
		FROM sessions
		WHERE id = ?
	`, sessionID)

	var s Session
	err := row.Scan(&s.ID, &s.Type, &s.Goal, &s.SuccessCriteria, &s.Boundaries, &s.Deliverables, &s.Status, &s.Metadata, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &s, nil
}

// GetOrCreateSession gets existing session or creates new one
func (m *Manager) GetOrCreateSession(sessionID, sessionType string) (*Session, error) {
	session, err := m.GetSession(sessionID)
	if err == nil {
		return session, nil
	}

	// Create new session
	newSession := &Session{
		ID:     sessionID,
		Type:   sessionType,
		Status: "active",
	}

	if err := m.CreateSession(newSession); err != nil {
		return nil, err
	}

	return newSession, nil
}

// UpdateSession updates session fields
func (m *Manager) UpdateSession(sessionID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := "UPDATE sessions SET "
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
	args = append(args, sessionID)

	_, err := m.db.Exec(query, args...)
	return err
}
