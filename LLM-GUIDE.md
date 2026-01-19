# llm-todo Quick Reference

## Commands You'll Need

### Create Tasks
```bash
llmtodo quick "Fix bug" "Write tests" "Deploy"    # 1-5 tasks
llmtodo add "Fix bug" "Write tests"                # Alias for quick
llmtodo create "Setup DB" "Run tests"              # Alias for quick
llmtodo import tasks.yaml                          # 5+ tasks (bulk)
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
llmtodo next              # Show next task
llmtodo done 1,2,3        # Complete tasks (comma-separated IDs)
llmtodo status            # Progress summary
```

### Query Tasks
```bash
llmtodo get p0            # High-priority tasks
llmtodo get pending       # All pending
llmtodo list              # Alias for get pending
llmtodo ls                # Alias for get pending
llmtodo extract           # Alias for get pending
llmtodo analyze p0        # Alias for get p0
llmtodo search "keyword"  # Search all fields
llmtodo find "keyword"    # Alias for search
llmtodo show 5            # Full details for task 5
llmtodo fetch 5           # Alias for show (MCP-style)
llmtodo read 5            # Alias for show
llmtodo view 5            # Alias for show
```

### Modify Tasks
```bash
llmtodo block 5 "waiting on PR"     # Mark blocked
llmtodo note 4 "needs review"       # Add note
llmtodo priority 3 150              # Change order
```

## Common Mistakes (Most Work!)

| You'll try | Status | Maps to | Notes |
|-----------|--------|---------|-------|
| `llmtodo add "task"` | ✅ Works | `quick` | Shows tip about import for bulk |
| `llmtodo create "task"` | ✅ Works | `quick` | |
| `llmtodo list` | ✅ Works | `get pending` | |
| `llmtodo ls` | ✅ Works | `get pending` | |
| `llmtodo fetch 5` | ✅ Works | `show` | MCP-style |
| `llmtodo find "word"` | ✅ Works | `search` | MCP-style |
| `llmtodo read 5` | ✅ Works | `show` | |
| `llmtodo view 5` | ✅ Works | `show` | |
| `llmtodo extract` | ✅ Works | `get pending` | Close enough |
| `llmtodo analyze` | ✅ Works | `get pending` | Close enough |
| `llmtodo update 5` | ⚠️ Redirect | Use `done`, `block`, or `note` | Shows helpful message |
| `llmtodo delete 5` | ⚠️ Redirect | Use `done` | Shows helpful message |

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
llmtodo quick "Setup database" "Write migrations" "Add tests"

# See what to do
llmtodo next

# Complete task
llmtodo done 1

# Check progress
llmtodo status
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
llmtodo import tasks.yaml

# Work
llmtodo next
llmtodo done 1
llmtodo status
```

### Multi-Session Work
```bash
# Session 1: Create and work
llmtodo import project-tasks.yaml
llmtodo next
llmtodo done 1,2,3

# Session 2: Resume (days later)
llmtodo status              # See where you left off
llmtodo get pending         # See remaining tasks
llmtodo search "database"   # Find specific work
llmtodo next                # Continue
```

### Batch Operations
```bash
# Complete multiple tasks at once
llmtodo get pending         # See task IDs
llmtodo done 1,3,5,7,9      # Complete scattered tasks

# Add notes to multiple
llmtodo note 10,11,12 "needs review"

# Block multiple
llmtodo block 2,4,6 "waiting on API"
```

## Session Management
```bash
llmtodo sessions            # List all projects
llmtodo use project-name    # Switch project
```

Sessions are auto-detected from directory name or can be set explicitly.

## File Locations

- Database: `~/.llm-todo/tasks.db`
- Templates: `~/.llm-todo/templates/` (P2 feature)
- Current session: `~/.llm-todo/current`

## Tips

1. **Write YAML in conversation:** You don't need a real file. Use `cat > file.yaml << 'EOF'` then import.
2. **Batch everything:** Completing 5 tasks = `done 1,2,3,4,5` not 5 separate commands.
3. **Search includes completed:** `llmtodo search "auth"` searches everything (for finding past work).
4. **Sessions persist:** State survives context loss. Pick up where you left off.
5. **Use get for scanning:** `llmtodo get pending` is minimal output. Use `llmtodo show 5` for details.

## When LLM Commands Fail

Most natural commands work via aliases. If something doesn't work:
1. Try `llmtodo --help` to see all commands
2. Check this guide for the right syntax
3. Error messages will suggest the correct command

## Quick Command Reference

**Most used:**
- `llmtodo next` - what to do
- `llmtodo done 1,2,3` - mark complete
- `llmtodo status` - check progress
- `llmtodo import file.yaml` - bulk create

**Query:**
- `llmtodo get p0` - high-priority
- `llmtodo get pending` - all pending
- `llmtodo search "word"` - find tasks

**Modify:**
- `llmtodo block 5 "reason"` - mark blocked
- `llmtodo note 5 "text"` - add note

That's it. Simple, fast, persistent.
