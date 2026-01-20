package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "default"
	}
	return wd
}

// getSessionID gets session ID from:
// 1. TODO_SESSION environment variable (for testing/override)
// 2. Active session from ~/.llm-todo/current
// 3. Current directory name
func getSessionID() string {
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

// printGoalSuggestion prints a helpful message suggesting to add session context
func printGoalSuggestion(sessionID string) {
	println()
	println("💡 Tip: Add session context for future cold-start sessions")
	println("   llmtodo session goal \"" + sessionID + "\" \"Brief description of what we're building/fixing\"")
	println()
	println("   Why? When you (or an LLM) returns to this project weeks later,")
	println("   session context helps understand what these tasks are for.")
	println("   Example: \"Refactor auth system to support OAuth2\"")
	println()
}
