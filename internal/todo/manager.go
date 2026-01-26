package todo

import (
	"database/sql"
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

// BeginTransaction starts a new transaction
func (m *Manager) BeginTransaction() (*sql.Tx, error) {
	return m.db.Begin()
}

// WithTransaction executes a function within a transaction
// If the function returns an error, the transaction is rolled back
// Otherwise it is committed
func (m *Manager) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
