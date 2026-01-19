package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/dtnitsch/llm-todo/internal/templates"
	"github.com/dtnitsch/llm-todo/internal/todo"
)

func addTemplateCommands(root *cobra.Command) {
	root.AddCommand(templatesCmd())
	root.AddCommand(templateCmd())
}

func templatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List available templates",
		Long: `List all available task templates.

Templates can be:
- Built-in (embedded in llm-todo)
- User-defined (~/.llm-todo/templates/)

Examples:
  todo templates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tmplList, err := templates.ListTemplates()
			if err != nil {
				return err
			}

			if len(tmplList) == 0 {
				fmt.Println("No templates available.")
				fmt.Println("\nCreate templates in: ~/.llm-todo/templates/")
				return nil
			}

			fmt.Printf("Available templates (%d total):\n\n", len(tmplList))

			// Group by source
			embedded := []templates.TemplateInfo{}
			user := []templates.TemplateInfo{}

			for _, tmpl := range tmplList {
				if tmpl.Source == "embedded" {
					embedded = append(embedded, tmpl)
				} else {
					user = append(user, tmpl)
				}
			}

			// Show embedded templates
			if len(embedded) > 0 {
				fmt.Println("Built-in templates:")
				for _, tmpl := range embedded {
					fmt.Printf("  %s (%d tasks)\n", tmpl.Name, tmpl.TaskCount)
					if tmpl.Description != "" {
						fmt.Printf("    %s\n", tmpl.Description)
					}
				}
				fmt.Println()
			}

			// Show user templates
			if len(user) > 0 {
				fmt.Println("User templates:")
				for _, tmpl := range user {
					fmt.Printf("  %s (%d tasks)\n", tmpl.Name, tmpl.TaskCount)
					if tmpl.Description != "" {
						fmt.Printf("    %s\n", tmpl.Description)
					}
				}
				fmt.Println()
			}

			fmt.Println("Usage:")
			fmt.Println("  todo template show <name>     # Preview template")
			fmt.Println("  todo template apply <name>    # Create tasks from template")

			return nil
		},
	}
}

func templateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Work with templates",
		Long: `Work with task templates.

Subcommands:
  show <name>     Preview template
  apply <name>    Create tasks from template`,
	}

	cmd.AddCommand(templateShowCmd())
	cmd.AddCommand(templateApplyCmd())

	return cmd
}

func templateShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <template-name>",
		Short: "Preview template details",
		Args:  cobra.ExactArgs(1),
		Long: `Show template details before applying.

Examples:
  todo template show vibe-crud-db
  todo template show my-api-pattern`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			tmpl, err := templates.LoadTemplate(name)
			if err != nil {
				return fmt.Errorf("failed to load template: %w", err)
			}

			fmt.Printf("Template: %s\n", name)
			if tmpl.Description != "" {
				fmt.Printf("Description: %s\n", tmpl.Description)
			}
			fmt.Printf("Tasks: %d\n\n", len(tmpl.Tasks))

			for i, task := range tmpl.Tasks {
				fmt.Printf("%d. %s\n", i+1, task.Title)
				if task.Priority != "" {
					fmt.Printf("   Priority: %s", task.Priority)
					if task.Effort != "" {
						fmt.Printf(", Effort: %s", task.Effort)
					}
					fmt.Println()
				}
				if len(task.Files) > 0 {
					fmt.Printf("   Files: %v\n", task.Files)
				}
				if len(task.DependsOn) > 0 {
					fmt.Printf("   Depends on: %v\n", task.DependsOn)
				}
				if len(task.Instructions) > 0 {
					if mustDo, exists := task.Instructions["must_do"]; exists && len(mustDo) > 0 {
						fmt.Printf("   Must do:\n")
						for _, item := range mustDo {
							fmt.Printf("     - %s\n", item)
						}
					}
					if mustNotDo, exists := task.Instructions["must_not_do"]; exists && len(mustNotDo) > 0 {
						fmt.Printf("   Must NOT do:\n")
						for _, item := range mustNotDo {
							fmt.Printf("     - %s\n", item)
						}
					}
				}
				fmt.Println()
			}

			fmt.Println("To apply this template:")
			fmt.Printf("  todo template apply %s\n", name)

			return nil
		},
	}
}

func templateApplyCmd() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "apply <template-name>",
		Short: "Create tasks from template",
		Args:  cobra.ExactArgs(1),
		Long: `Apply a template to create tasks in current session.

Examples:
  todo template apply vibe-crud-db
  todo template apply vibe-crud-db --domain posts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			sessionID := getSessionID()

			mgr, err := todo.NewManager("")
			if err != nil {
				return err
			}
			defer mgr.Close()

			// Variable substitution
			vars := make(map[string]string)
			if domain != "" {
				vars["domain"] = domain
			}

			count, err := templates.ApplyTemplate(mgr, sessionID, name, vars)
			if err != nil {
				return err
			}

			fmt.Printf("✓ Created %d tasks from template: %s\n", count, name)
			fmt.Println("\nNext steps:")
			fmt.Println("  todo next       # Start working")
			fmt.Println("  todo status     # Check progress")

			return nil
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain name for variable substitution (e.g., 'posts')")

	return cmd
}
