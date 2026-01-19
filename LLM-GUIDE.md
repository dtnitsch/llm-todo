# llm-todo Quick Reference

## Commands You'll Need

### Create Tasks
```bash
todo quick "Fix bug" "Write tests" "Deploy"    # 1-5 tasks
todo add "Fix bug" "Write tests"                # Alias for quick
todo create "Setup DB" "Run tests"              # Alias for quick
todo import tasks.yaml                          # 5+ tasks (bulk)
```

**YAML format:**
```yaml
- title: "Task description"
  priority: p0
  effort: m
  files: ["src/main.go", "tests/main_test.go"]
  instructions:
    must_do:
      - "Add validation"
      - "Include error handling"
    must_not_do:
      - "Don't skip tests"
```

### Work on Tasks
```bash
todo next              # Show next task
todo done 1,2,3        # Complete tasks (comma-separated IDs)
todo status            # Progress summary
```

### Query Tasks
```bash
todo get p0            # High-priority tasks
todo get pending       # All pending
todo list              # Alias for get pending
todo ls                # Alias for get pending
todo extract           # Alias for get pending
todo analyze p0        # Alias for get p0
todo search "keyword"  # Search all fields
todo find "keyword"    # Alias for search
todo show 5            # Full details for task 5
todo fetch 5           # Alias for show (MCP-style)
todo read 5            # Alias for show
todo view 5            # Alias for show
```

### Modify Tasks
```bash
todo block 5 "waiting on PR"     # Mark blocked
todo note 4 "needs review"       # Add note
todo priority 3 150              # Change order
```

## Common Mistakes (Most Work!)

| You'll try | Status | Maps to | Notes |
|-----------|--------|---------|-------|
| `todo add "task"` | ✅ Works | `quick` | Shows tip about import for bulk |
| `todo create "task"` | ✅ Works | `quick` | |
| `todo list` | ✅ Works | `get pending` | |
| `todo ls` | ✅ Works | `get pending` | |
| `todo fetch 5` | ✅ Works | `show` | MCP-style |
| `todo find "word"` | ✅ Works | `search` | MCP-style |
| `todo read 5` | ✅ Works | `show` | |
| `todo view 5` | ✅ Works | `show` | |
| `todo extract` | ✅ Works | `get pending` | Close enough |
| `todo analyze` | ✅ Works | `get pending` | Close enough |
| `todo update 5` | ⚠️ Redirect | Use `done`, `block`, or `note` | Shows helpful message |
| `todo delete 5` | ⚠️ Redirect | Use `done` | Shows helpful message |

## Rules

1. **1-5 tasks:** Use `quick` or `add`
2. **5+ tasks:** Use `import` (write YAML → save → import)
3. **Batch operations:** Comma-separated IDs (`done 1,2,3` not `done 1; done 2; done 3`)
4. **Output formats:**
   - `get` = minimal (IDs + titles)
   - `show` = full details
   - `next` = next task with instructions

## Workflow Examples

### Quick Session (1-5 Tasks)
```bash
# Create tasks
todo quick "Setup database" "Write migrations" "Add tests"

# See what to do
todo next

# Complete task
todo done 1

# Check progress
todo status
```

### Bulk Import (5+ Tasks)
```bash
# Write YAML in conversation
cat > tasks.yaml << 'EOF'
- title: "Setup PostgreSQL database"
  priority: p0
  files: ["db/schema.sql"]
- title: "Implement auth API"
  priority: p0
  files: ["api/auth.go"]
  instructions:
    must_do:
      - "POST /auth/login"
      - "JWT token generation"
- title: "Write integration tests"
  priority: p1
EOF

# Import
todo import tasks.yaml

# Work
todo next
todo done 1
todo status
```

### Multi-Session Work
```bash
# Session 1: Create and work
todo import project-tasks.yaml
todo next
todo done 1,2,3

# Session 2: Resume (days later)
todo status              # See where you left off
todo get pending         # See remaining tasks
todo search "database"   # Find specific work
todo next                # Continue
```

### Batch Operations
```bash
# Complete multiple tasks at once
todo get pending         # See task IDs
todo done 1,3,5,7,9      # Complete scattered tasks

# Add notes to multiple
todo note 10,11,12 "needs review"

# Block multiple
todo block 2,4,6 "waiting on API"
```

## Session Management
```bash
todo sessions            # List all projects
todo use project-name    # Switch project
```

Sessions are auto-detected from directory name or can be set explicitly.

## File Locations

- Database: `~/.llm-todo/tasks.db`
- Templates: `~/.llm-todo/templates/` (P2 feature)
- Current session: `~/.llm-todo/current`

## Tips

1. **Write YAML in conversation:** You don't need a real file. Use `cat > file.yaml << 'EOF'` then import.
2. **Batch everything:** Completing 5 tasks = `done 1,2,3,4,5` not 5 separate commands.
3. **Search includes completed:** `todo search "auth"` searches everything (for finding past work).
4. **Sessions persist:** State survives context loss. Pick up where you left off.
5. **Use get for scanning:** `todo get pending` is minimal output. Use `todo show 5` for details.

## When LLM Commands Fail

Most natural commands work via aliases. If something doesn't work:
1. Try `todo --help` to see all commands
2. Check this guide for the right syntax
3. Error messages will suggest the correct command

## Quick Command Reference

**Most used:**
- `todo next` - what to do
- `todo done 1,2,3` - mark complete
- `todo status` - check progress
- `todo import file.yaml` - bulk create

**Query:**
- `todo get p0` - high-priority
- `todo get pending` - all pending
- `todo search "word"` - find tasks

**Modify:**
- `todo block 5 "reason"` - mark blocked
- `todo note 5 "text"` - add note

That's it. Simple, fast, persistent.
