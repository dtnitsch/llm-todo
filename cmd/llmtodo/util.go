package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "default"
	}
	return wd
}

// generateSessionID creates a session ID from directory and optional name
// If name is provided: {directory}-{name}
// If name is empty: {directory}-{timestamp}
func generateSessionID(name string) string {
	dirName := filepath.Base(mustGetwd())

	if name != "" {
		return dirName + "-" + name
	}

	// Generate timestamp: YYYYMMDD-HHMM
	timestamp := time.Now().Format("20060102-1504")
	return dirName + "-" + timestamp
}

// getSessionID gets session ID from:
// 1. TODO_SESSION environment variable (for testing/override)
// 2. Active session from ~/.llm-todo/current
// 3. Current directory name
func getSessionID() string {
	return getSessionIDWithOverride("")
}

// getSessionIDWithOverride gets session ID with optional override
// Priority order:
// 1. Override parameter (from --session flag)
// 2. TODO_SESSION environment variable (for testing/override)
// 3. Active session from ~/.llm-todo/current
// 4. Current directory name
func getSessionIDWithOverride(override string) string {
	// Check for flag override first
	if override != "" {
		return override
	}

	// Check for environment override (for testing)
	if envSession := os.Getenv("TODO_SESSION"); envSession != "" {
		return envSession
	}

	// Check if active session is set
	currentSession, err := todo.GetCurrentSession()
	if err == nil && currentSession != "" {
		return currentSession
	}

	// Default to directory name
	wd := mustGetwd()
	return filepath.Base(wd)
}

func generateActiveForm(title string) string {
	verbs := map[string]string{
		"Create":    "Creating",
		"Implement": "Implementing",
		"Build":     "Building",
		"Fix":       "Fixing",
		"Update":    "Updating",
		"Add":       "Adding",
		"Remove":    "Removing",
		"Refactor":  "Refactoring",
		"Test":      "Testing",
		"Deploy":    "Deploying",
	}

	for verb, ing := range verbs {
		if strings.HasPrefix(title, verb+" ") {
			return ing + strings.TrimPrefix(title, verb)
		}
	}

	return "Working on: " + title
}

// printNamingHint prints a suggestion to use --name for clarity
func printNamingHint() {
	println()
	println("TIP: Use --name for clarity")
	println("  llmtodo quick \"task1\" \"task2\" --name feature-name")
	println()
}

// printGoalSuggestion prints a helpful message suggesting to add session context
func printGoalSuggestion(sessionID string) {
	println()
	println("RECOMMEND: Add session goal for cold-start orientation")
	println("  llmtodo session goal " + sessionID + " \"What we're building\"")
	println()
	println("Session goal helps LLMs orient without conversation history.")
	println("Example: llmtodo session goal " + sessionID + " \"Refactor auth system to support OAuth2\"")
	println()
}
