package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	path string
}

// Open opens or creates a SQLite database at the given path
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &DB{DB: sqlDB, path: path}, nil
}

// Init initializes the database schema
func (db *DB) Init() error {
	version, err := db.getSchemaVersion()
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	if version == SchemaVersion {
		return nil
	}

	if version == 0 {
		return db.initializeSchema()
	}

	return fmt.Errorf("unknown schema version %d", version)
}

// WithTransaction executes a function within a transaction
func (db *DB) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) getSchemaVersion() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&count)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	var version int
	err = db.QueryRow("SELECT version FROM schema_version ORDER BY applied_at DESC LIMIT 1").Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (db *DB) initializeSchema() error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		if _, err := tx.Exec(schemaSQL); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}

		if _, err := tx.Exec("INSERT INTO schema_version (version, applied_at) VALUES (?, datetime('now'))", SchemaVersion); err != nil {
			return fmt.Errorf("failed to set schema version: %w", err)
		}

		return nil
	})
}

// DefaultPath returns the default database path
func DefaultPath() string {
	// Check for environment override (for testing)
	if envPath := os.Getenv("TODO_DB"); envPath != "" {
		return envPath
	}

	// Check for project-local database first
	if _, err := os.Stat(".llm-todo/tasks.db"); err == nil {
		return ".llm-todo/tasks.db"
	}

	// Fall back to global database
	home, err := os.UserHomeDir()
	if err != nil {
		return ".llm-todo/tasks.db"
	}

	return filepath.Join(home, ".llm-todo", "tasks.db")
}
