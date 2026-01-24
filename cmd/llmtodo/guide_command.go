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
		Short: "LLM quick reference for token-efficient workflows",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(`LLM QUICK REFERENCE - Token-Efficient Workflows

CREATE SESSION (once per project):
  llmtodo quick "task1" "task2" --name feature-name  # Named session
  llmtodo quick "task1" "task2"                      # Auto: {dir}-{timestamp}
  llmtodo import tasks.yaml                          # Bulk import (5+ tasks)

WORK LOOP (repeat):
  llmtodo next                             # Shows next task (60-100 tokens)
  [do the work]
  llmtodo done                             # Complete current task

QUERY (when needed):
  llmtodo get p0                           # IDs + titles only (60 tokens)
  llmtodo show 5                           # Full details (500 tokens)

BATCH OPERATIONS (token saver):
  llmtodo done 1,2,3                       # 1 command vs 3 (95% token savings)
  llmtodo block 4,5 "waiting on PR"        # Block multiple tasks
  llmtodo note 6,7 "needs review"          # Add notes to multiple

CONTEXT SWITCHING:
  llmtodo sessions                         # List all active sessions
  llmtodo use other-project                # Switch current session
  llmtodo get p0 --session other           # Query without switching

ENRICHMENT (for context after session loss):
  llmtodo enrich 5                         # Add files/instructions/context
  llmtodo next                             # Shows warning if context missing

TOKEN EFFICIENCY TIPS:
  - Use 'get' before 'show' (88% token savings: 60 vs 500)
  - Batch operations save 95% tokens (done 1,2,3 vs 3 separate commands)
  - Session switching: use 'use' once or '--session' flag for queries
  - 'next' shows enrichment inline (avoids extra 'show' calls)

COMMON WORKFLOWS:

Bug fix workflow:
  llmtodo quick "Reproduce bug" "Fix" "Test" "PR"
  llmtodo next           # Get first task
  llmtodo done           # Complete when done
  llmtodo next           # Next task

Feature workflow:
  llmtodo code "Design" "Implement" "Test" "Docs" "Deploy"
  llmtodo enrich 1,2,3   # Add context upfront
  llmtodo get p0         # See all high-priority
  llmtodo done 1,2       # Complete multiple

Multi-session workflow:
  llmtodo sessions                        # See all active
  llmtodo get p0 --session project-a      # Check project-a
  llmtodo get p0 --session project-b      # Check project-b
  llmtodo use project-a                   # Switch to work on it

Full command reference: llmtodo --help
`)
		},
	}
}
