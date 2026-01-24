package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addHelpfulAliases registers natural command aliases that LLMs will try
func addHelpfulAliases(root *cobra.Command) {
	// Working aliases - natural language
	root.AddCommand(addCmd())
	root.AddCommand(createCmd())
	root.AddCommand(listCmd())
	root.AddCommand(lsCmd())

	// MCP-inspired aliases
	root.AddCommand(fetchCmd())
	root.AddCommand(findCmd())
	root.AddCommand(readCmd())
	root.AddCommand(viewCmd())
	root.AddCommand(extractCmd())
	root.AddCommand(analyzeCmd())

	// Export/template aliases
	root.AddCommand(dumpCmd())
	root.AddCommand(saveCmd())
	root.AddCommand(backupCmd())
	root.AddCommand(importTemplateCmd())
	root.AddCommand(formatCmd())
	root.AddCommand(exampleCmd())
	root.AddCommand(scaffoldCmd())
	root.AddCommand(initCmd())
	root.AddCommand(setupCmd())

	// Helpful redirects (commands that don't map directly)
	root.AddCommand(updateHelpCmd())
	root.AddCommand(deleteHelpCmd())
}

// addCmd - "add" works as alias to "quick" with helpful tip
func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <task1> <task2> ...",
		Short: "Create tasks (alias for 'quick')",
		Long: `Create tasks in current session (alias for 'quick').

Examples:
  todo add "Fix login bug" "Write tests" "Deploy to staging"

💡 Tip: For 5+ tasks, use 'todo import tasks.yaml' (bulk creation)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show tip on first use
			fmt.Println("💡 Tip: 'add' is an alias. For 5+ tasks, use 'todo import tasks.yaml'")

			// Reuse quick command logic
			return quickCmd().RunE(cmd, args)
		},
	}
	return cmd
}

// createCmd - "create" works as alias to "quick"
func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <task1> <task2> ...",
		Short: "Create tasks (alias for 'quick')",
		Long: `Create tasks in current session (alias for 'quick').

Examples:
  todo create "Setup database" "Write migrations"

💡 Tip: For bulk creation, use 'todo import tasks.yaml'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return quickCmd().RunE(cmd, args)
		},
	}
	return cmd
}

// listCmd - "list" works as alias to "get pending"
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pending tasks (alias for 'get pending')",
		Long: `List all pending tasks (alias for 'get pending').

Output: Minimal (IDs + titles)
Full details: todo show <id>

Other filters:
  todo get p0         # High-priority only
  todo get completed  # Completed tasks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Call get pending
			return getCmd().RunE(cmd, []string{"pending"})
		},
	}
}

// lsCmd - "ls" works as alias to "get pending"
func lsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List pending tasks (alias for 'get pending')",
		Long: `List all pending tasks (alias for 'get pending').

Common CLI pattern: 'ls' to see what's there.

Output: Minimal (IDs + titles)
Full details: todo show <id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getCmd().RunE(cmd, []string{"pending"})
		},
	}
}

// updateHelpCmd - "update" doesn't map directly, show helpful redirect
func updateHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Use specific commands instead",
		Long: `You tried: todo update

Use specific commands:

  todo done 1,2,3                # Mark complete
  todo block 5 "waiting on PR"   # Mark blocked
  todo note 4 "needs testing"    # Add note
  todo priority 3 150            # Change order

Batch syntax: Comma-separated IDs (1,2,3)

Example workflow:
  todo get pending    # See task IDs
  todo done 1,3,5     # Complete multiple
  todo status         # Check progress`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Long)
		},
	}
}

// deleteHelpCmd - "delete" redirects to "done" (we don't delete by default)
func deleteHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Use 'done' to mark tasks complete",
		Long: `You tried: todo delete

llm-todo keeps completed tasks for search history.

To mark as completed (recommended):
  todo done 1,2,3

Example workflow:
  todo get pending    # Find task ID
  todo done 5         # Mark complete
  todo search "auth"  # Search includes completed tasks

Note: Permanent deletion is not yet implemented.
For now: Use 'done' to mark complete.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Long)
		},
	}
}

// MCP-inspired aliases

// fetchCmd - "fetch" as alias to "show" (MCP pattern: fetch resource by ID)
func fetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch <task-id>",
		Short: "Fetch task details (MCP-style alias for 'show')",
		Long: `Fetch task details by ID (alias for 'show').

MCP-style command for getting a specific resource.

Examples:
  todo fetch 5        # Get full details for task 5
  todo fetch 12       # Get full details for task 12

Output: Full task details (title, status, priority, files, instructions)
Minimal output: todo get pending`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCmd().RunE(cmd, args)
		},
	}
}

// findCmd - "find" as alias to "search"
func findCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find <keyword>",
		Short: "Find tasks by keyword (alias for 'search')",
		Long: `Find tasks matching keyword (alias for 'search').

Searches: titles, notes, files, instructions, refs

Examples:
  todo find "auth"        # Find auth-related tasks
  todo find "database"    # Find database tasks

Includes completed tasks by default.
Pending only: todo find "auth" --pending`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return searchCmd().RunE(cmd, args)
		},
	}
}

// readCmd - "read" as alias to "show"
func readCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <task-id>",
		Short: "Read task details (alias for 'show')",
		Long: `Read task details by ID (alias for 'show').

Examples:
  todo read 5         # Read task 5
  todo read 12        # Read task 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCmd().RunE(cmd, args)
		},
	}
}

// viewCmd - "view" as alias to "show"
func viewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view <task-id>",
		Short: "View task details (alias for 'show')",
		Long: `View task details by ID (alias for 'show').

Examples:
  todo view 5         # View task 5
  todo view 12        # View task 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showCmd().RunE(cmd, args)
		},
	}
}

// extractCmd - "extract" as alias to "get pending" (close enough for now)
func extractCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extract [priority|status]",
		Short: "Extract/list tasks (alias for 'get')",
		Long: `Extract/list tasks (alias for 'get pending').

Shows minimal output (IDs + titles) for further processing.

Examples:
  todo extract              # All pending tasks
  todo extract p0           # High-priority tasks
  todo extract completed    # Completed tasks

💡 "Extract" maps to 'get' - shows you the task list to extract from.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to pending if no args
			if len(args) == 0 {
				args = []string{"pending"}
			}
			return getCmd().RunE(cmd, args)
		},
	}
}

// analyzeCmd - "analyze" as alias to "get pending" (see what's there to analyze)
func analyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [priority|status]",
		Short: "Analyze/list tasks (alias for 'get')",
		Long: `Analyze tasks (alias for 'get pending').

Shows minimal output (IDs + titles) so you can analyze the task list.

Examples:
  todo analyze              # All pending tasks
  todo analyze p0           # Analyze high-priority
  todo analyze completed    # Analyze completed work

💡 "Analyze" maps to 'get' - shows you the task list to analyze.
For detailed analysis: todo status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to pending if no args
			if len(args) == 0 {
				args = []string{"pending"}
			}
			return getCmd().RunE(cmd, args)
		},
	}
}

// Export/Template aliases

// dumpCmd - "dump" as alias to "export" (MCP pattern)
func dumpCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "dump <file.yaml>",
		Short:  "Dump tasks to file (alias for 'export')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportCmd().RunE(cmd, args)
		},
	}
}

// saveCmd - "save" as alias to "export"
func saveCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "save <file.yaml>",
		Short:  "Save tasks to file (alias for 'export')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportCmd().RunE(cmd, args)
		},
	}
}

// backupCmd - "backup" as alias to "export"
func backupCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "backup <file.yaml>",
		Short:  "Backup tasks to file (alias for 'export')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportCmd().RunE(cmd, args)
		},
	}
}

// templateCmd - "template" as standalone command for import --template
func importTemplateCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "template",
		Short:  "Show import template (alias for 'import --template')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("template", "true")
			return importCmd().RunE(cmd, []string{})
		},
	}
}

// formatCmd - "format" as alias to "import --template"
func formatCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "format",
		Short:  "Show format template (alias for 'import --template')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("template", "true")
			return importCmd().RunE(cmd, []string{})
		},
	}
}

// exampleCmd - "example" as alias to "import --template"
func exampleCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "example",
		Short:  "Show example template (alias for 'import --template')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("template", "true")
			return importCmd().RunE(cmd, []string{})
		},
	}
}

// scaffoldCmd - "scaffold" as standalone for export --scaffold
func scaffoldCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "scaffold <dir>",
		Short:  "Create scaffold structure (alias for 'export --scaffold')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("scaffold", "true")
			return exportCmd().RunE(cmd, args)
		},
	}
}

// initCmd - "init" as alias to "export --scaffold"
func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "init <dir>",
		Short:  "Initialize structure (alias for 'export --scaffold')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("scaffold", "true")
			return exportCmd().RunE(cmd, args)
		},
	}
}

// setupCmd - "setup" as alias to "export --scaffold"
func setupCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "setup <dir>",
		Short:  "Setup structure (alias for 'export --scaffold')",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set("scaffold", "true")
			return exportCmd().RunE(cmd, args)
		},
	}
}
