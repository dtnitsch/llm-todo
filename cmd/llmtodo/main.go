package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "llmtodo",
	Short: "Persistent task management for LLM workflows",
	Long: `llm-todo - Token-efficient task management for LLMs

Built FOR LLMs, BY LLMs. Minimal commands, rich structured output.

TASK CREATION:
  quick        Create 1-5 tasks: llmtodo quick "task 1" "task 2"
               (aliases: add, create)
  import       Import from YAML: llmtodo import tasks.yaml
  code         Create code project session (20+ tasks)
  research     Create research project session

TASK WORKFLOW:
  next         Show next task with details
  done         Complete tasks: llmtodo done 1,2,3
  block        Block tasks: llmtodo block 5 "reason"
  note         Add notes: llmtodo note 5 "context"

TASK QUERYING:
  get          List tasks: llmtodo get p0|p1|pending|blocked
               (aliases: analyze, extract, list, ls)
  show         Full details: llmtodo show 5
               (aliases: fetch, read, view)
  search       Find tasks: llmtodo search "keyword"
               (aliases: find)
  status       Session progress summary

COMMON PATTERNS:
  llmtodo next              # See what to work on
  llmtodo done              # Complete current task (no ID needed)
  llmtodo get p0            # High-priority tasks
  llmtodo show 5            # Task details
  llmtodo search "auth"     # Find auth-related tasks

Use "llmtodo help <command>" for more information about a specific command.`,
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
	rootCmd.AddCommand(versionCmd())

	// Hide default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Use custom templates (shows Long description only, hides command list)
	rootCmd.SetHelpTemplate(`{{.Long}}
`)
	rootCmd.SetUsageTemplate(`{{.Long}}
`)
}
