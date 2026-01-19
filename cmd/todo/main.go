package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "Persistent task management for LLM workflows",
	Long: `llm-todo - Token-efficient task management for LLMs

Built FOR LLMs, BY LLMs. Minimal commands, rich structured output.

Examples:
  todo quick 'Fix bug' 'Update docs' 'Run tests'
  todo next
  todo done 3,5,7
  todo get p0`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	addModeCommands(rootCmd)
	addCoreCommands(rootCmd)
	addQueryCommands(rootCmd)
	addPriorityCommands(rootCmd)
	addTemplateCommands(rootCmd)
	addHelpfulAliases(rootCmd)
}
