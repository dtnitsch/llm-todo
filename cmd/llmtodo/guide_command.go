package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(guideCmd())
}

func guideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "LLM quick reference for common workflows",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(`LLM QUICK REFERENCE

BASIC WORKFLOW:
  llmtodo quick "task1" "task2" --name feature-x
  llmtodo next
  llmtodo done
  llmtodo next

QUERY:
  llmtodo get p0           # Minimal: IDs + titles
  llmtodo show 5           # Full: all fields
  llmtodo search "auth"

BATCH:
  llmtodo done 1,2,3
  llmtodo block 4,5 "waiting on PR"

SESSIONS:
  llmtodo sessions → llmtodo use {other-session}

ENRICHMENT (mid-stream when you have context):
  1. llmtodo code "task1" "task2" "task3"
  2. Read: ~/.llm-todo/enrichment/{session}.yaml
  3. Write: Overwrite entire file with enriched version
  4. llmtodo import ~/.llm-todo/enrichment/{session}.yaml

  Hint: llmtodo enrich --status for suggestions

MEMORY EFFICIENCY:
  - Use 'get' before 'show' (88% smaller)
  - 'next' shows enrichment inline (no extra reads)

MULTI-SESSION:
  llmtodo get p0 --session project-a
  llmtodo get p0 --session project-b
  llmtodo use project-a

Full help: llmtodo --help
`)
		},
	}
}
