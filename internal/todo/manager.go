package todo

import (
	"fmt"

	"github.com/dtnitsch/llm-todo/internal/db"
)

type Manager struct {
	db *db.DB
}

// NewManager creates a new todo manager
func NewManager(dbPath string) (*Manager, error) {
	if dbPath == "" {
		dbPath = db.DefaultPath()
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := database.Init(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Manager{db: database}, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}
