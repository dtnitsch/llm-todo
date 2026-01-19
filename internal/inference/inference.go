package inference

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InferSessionID infers session ID from current directory
func InferSessionID() string {
	wd, err := os.Getwd()
	if err != nil {
		return "default"
	}
	return filepath.Base(wd)
}

// InferTaskType infers task type from description
func InferTaskType(description string) string {
	desc := strings.ToLower(description)

	// Research indicators
	if strings.Contains(desc, "http://") || strings.Contains(desc, "https://") {
		return "research"
	}
	if strings.Contains(desc, "analyze") || strings.Contains(desc, "research") {
		return "research"
	}

	// Code indicators
	if _, err := os.Stat(".git"); err == nil {
		return "task"
	}

	return "task"
}

// ExtractFilePaths extracts file paths from description
func ExtractFilePaths(description string) []string {
	// Match patterns like: file.go, path/to/file.go, file.go:123
	re := regexp.MustCompile(`([\w/.-]+\.\w+)(:\d+)?`)
	matches := re.FindAllString(description, -1)

	if len(matches) == 0 {
		return nil
	}

	// Deduplicate
	seen := make(map[string]bool)
	var files []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			files = append(files, match)
		}
	}

	return files
}

// InferEffort infers effort from task description length
func InferEffort(description string) string {
	words := len(strings.Fields(description))

	if words < 10 {
		return "xs"
	} else if words < 30 {
		return "s"
	}
	return "m"
}
