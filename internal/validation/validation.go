package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dtnitsch/llm-todo/internal/todo"
)

// CheckPrerequisites checks if all prerequisite tasks are completed
func CheckPrerequisites(mgr *todo.Manager, task *todo.Task) ([]string, error) {
	if task.DependantIDs == "" || task.DependantIDs == "[]" {
		return nil, nil
	}

	var depIDs []int
	if err := json.Unmarshal([]byte(task.DependantIDs), &depIDs); err != nil {
		return nil, err
	}

	var incomplete []string
	for _, depID := range depIDs {
		depTask, err := mgr.GetTask(depID)
		if err != nil {
			continue
		}
		if depTask.Status != "completed" {
			incomplete = append(incomplete, fmt.Sprintf("#%d (%s)", depID, depTask.Task))
		}
	}

	return incomplete, nil
}

// DetectDuplicates finds similar completed tasks (keyword overlap)
func DetectDuplicates(mgr *todo.Manager, sessionID, taskTitle string) ([]string, error) {
	completed, err := mgr.ListTasks(sessionID, map[string]string{"status": "completed"})
	if err != nil {
		return nil, err
	}

	keywords := extractKeywords(taskTitle)
	if len(keywords) == 0 {
		return nil, nil
	}

	var duplicates []string
	for _, task := range completed {
		overlap := calculateKeywordOverlap(keywords, extractKeywords(task.Task))
		if overlap > 0.6 { // 60% overlap threshold
			duplicates = append(duplicates, fmt.Sprintf("#%d: %s", task.ID, task.Task))
		}
	}

	return duplicates, nil
}

// FindUnblockedTasks finds tasks that can be unblocked after completing a task
func FindUnblockedTasks(mgr *todo.Manager, sessionID string, completedID int) ([]*todo.Task, error) {
	blocked, err := mgr.ListTasks(sessionID, map[string]string{"status": "blocked"})
	if err != nil {
		return nil, err
	}

	var unblocked []*todo.Task
	for _, task := range blocked {
		if task.DependantIDs == "" || task.DependantIDs == "[]" {
			continue
		}

		var depIDs []int
		if err := json.Unmarshal([]byte(task.DependantIDs), &depIDs); err != nil {
			continue
		}

		// Check if this task depended on the completed task
		dependedOnCompleted := false
		for _, depID := range depIDs {
			if depID == completedID {
				dependedOnCompleted = true
				break
			}
		}

		if !dependedOnCompleted {
			continue
		}

		// Check if all prerequisites are now complete
		incomplete, _ := CheckPrerequisites(mgr, task)
		if len(incomplete) == 0 {
			unblocked = append(unblocked, task)
		}
	}

	return unblocked, nil
}

func extractKeywords(text string) []string {
	// Remove common words and extract meaningful keywords
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "with": true, "from": true, "is": true, "are": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

func calculateKeywordOverlap(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	aSet := make(map[string]bool)
	for _, word := range a {
		aSet[word] = true
	}

	overlap := 0
	for _, word := range b {
		if aSet[word] {
			overlap++
		}
	}

	return float64(overlap) / float64(len(a))
}
