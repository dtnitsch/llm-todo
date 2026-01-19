package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// promptOptional prompts for input, allows empty (Enter to skip)
func promptOptional(prompt string) string {
	fmt.Printf("%s [Enter to skip]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptCodeSession prompts for code session metadata
func promptCodeSession() (goal, boundaries, successCriteria string) {
	goal = promptOptional("Goal")
	boundaries = promptOptional("Boundaries")
	successCriteria = promptOptional("Success criteria")
	return
}

// promptResearchSession prompts for research session metadata
func promptResearchSession() (goal, deliverables string) {
	goal = promptOptional("Goal")
	deliverables = promptOptional("Deliverables")
	return
}

// encodeJSONArray converts comma-separated string to JSON array
func encodeJSONArray(input string) string {
	if input == "" {
		return "[]"
	}

	parts := strings.Split(input, ",")
	var items []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}

	data, _ := json.Marshal(items)
	return string(data)
}
