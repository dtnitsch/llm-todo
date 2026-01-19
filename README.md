# llm-todo

Persistent task management for LLM workflows. Built FOR LLMs, BY LLMs.

**95-99% token savings** vs TodoWrite. Proven with empirical testing (see `test/TEST-RESULTS.md`).

## Installation

### Via go install (Recommended)

```bash
go install github.com/dtnitsch/llm-todo/cmd/todo@latest
```

### From source

```bash
git clone https://github.com/dtnitsch/llm-todo.git
cd llm-todo
go build -o ~/bin/todo ./cmd/todo
```

Ensure `~/bin` (or your chosen install location) is in your PATH.

## Quick Start

```bash
# Create quick session
todo quick "Fix bug" "Update docs" "Run tests"

# See next task
todo next

# Complete tasks (batch)
todo done 1,2,3

# List tasks (minimal - saves tokens)
todo get p0
todo get pending

# Show full details
todo show 3

# Check progress
todo status
```

## Features

- **Token efficient**: Minimal reads (500 tokens vs 2000), batch operations
- **Conditional output**: Only shows set fields, no null garbage
- **Three modes**: quick (3-5 tasks), code (20+ tasks), research (non-code)
- **Priority ordering**: Sub-order tasks within p0/p1/p2
- **Cross-LLM compatible**: Works with Claude, Gemini, ChatGPT

## Commands

### Create Sessions
- `todo quick <tasks...>` - Quick session (3-5 tasks)
- `todo code <tasks...>` - Code project (20+ tasks)
- `todo research <tasks...>` - Research project

### Read Tasks
- `todo get p0` - Minimal list (IDs + titles)
- `todo get pending` - All pending tasks
- `todo show 3` - Full details for task #3
- `todo next` - Next task with full details
- `todo` or `todo status` - Progress summary

### Update Tasks
- `todo done` - Complete current task
- `todo done 3,5,7` - Batch complete
- `todo block 2,4 "reason"` - Batch block
- `todo note 3 "context"` - Add note
- `todo priority 4 50` - Reorder task

## Database

- Global: `~/.llm-todo/tasks.db`
- Project: `.llm-todo/tasks.db` (if exists)

## Architecture

Clean separation, no files over 500 lines:

```
cmd/todo/          # CLI commands
internal/db/       # Database layer
internal/todo/     # Core logic
internal/formatter/# Conditional output
internal/inference/# Auto-detection
```

Total: ~3000 lines across 20 files
