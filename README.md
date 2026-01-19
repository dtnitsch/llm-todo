# llm-todo

Persistent task management for LLM workflows. Built FOR LLMs, BY LLMs.

**95-99% token savings** vs TodoWrite. Proven with empirical testing (see `test/TEST-RESULTS.md`).

## Installation

### Via go install (Recommended)

```bash
go install github.com/dtnitsch/llm-todo/cmd/llmtodo@latest
```

### From source

```bash
git clone https://github.com/dtnitsch/llm-todo.git
cd llm-todo
go build -o ~/bin/llmtodo ./cmd/llmtodo
```

Ensure `~/bin` (or your chosen install location) is in your PATH.

## Quick Start

```bash
# Create quick session
llmtodo quick "Fix bug" "Update docs" "Run tests"

# See next task
llmtodo next

# Complete tasks (batch)
llmtodo done 1,2,3

# List tasks (minimal - saves tokens)
llmtodo get p0
llmtodo get pending

# Show full details
llmtodo show 3

# Check progress
llmtodo status
```

## Features

- **Token efficient**: Minimal reads (500 tokens vs 2000), batch operations
- **Conditional output**: Only shows set fields, no null garbage
- **Three modes**: quick (3-5 tasks), code (20+ tasks), research (non-code)
- **Priority ordering**: Sub-order tasks within p0/p1/p2
- **Cross-LLM compatible**: Works with Claude, Gemini, ChatGPT

## Commands

### Create Sessions
- `llmtodo quick <tasks...>` - Quick session (3-5 tasks)
- `llmtodo code <tasks...>` - Code project (20+ tasks)
- `llmtodo research <tasks...>` - Research project

### Read Tasks
- `llmtodo get p0` - Minimal list (IDs + titles)
- `llmtodo get pending` - All pending tasks
- `llmtodo show 3` - Full details for task #3
- `llmtodo next` - Next task with full details
- `llmtodo` or `llmtodo status` - Progress summary

### Update Tasks
- `llmtodo done` - Complete current task
- `llmtodo done 3,5,7` - Batch complete
- `llmtodo block 2,4 "reason"` - Batch block
- `llmtodo note 3 "context"` - Add note
- `llmtodo priority 4 50` - Reorder task

## Database

- Global: `~/.llm-todo/tasks.db`
- Project: `.llm-todo/tasks.db` (if exists)

## Architecture

Clean separation, no files over 500 lines:

```
cmd/llmtodo/          # CLI commands
internal/db/       # Database layer
internal/todo/     # Core logic
internal/formatter/# Conditional output
internal/inference/# Auto-detection
```

Total: ~3000 lines across 20 files
