# llm-todo Project Notes

## This Project Uses llm-todo (Dogfooding)

This project tracks its own tasks using llm-todo. Use these commands:

```bash
# See what to work on
llmtodo next

# Check progress
llmtodo status

# Complete tasks
llmtodo done 1,2,3

# Search for specific work
llmtodo search "template"
llmtodo find "alias"
```

## Current Session

Session: **llm-todo** (auto-detected from directory name)

Database: `~/.llm-todo/tasks.db`

## Quick Reference

Full LLM guide: `cat LLM-GUIDE.md`

All aliases work (add, create, list, fetch, find, etc.)

## Project Status

- P0+P1: ✅ Complete (28/52 tasks done)
- P2: Templates (pending)
- Test results: `test/TEST-RESULTS.md`
- Implementation summary: `IMPLEMENTATION.md`

## Key Files

- `cmd/todo/` - CLI commands
- `internal/` - Core logic (todo, db, formatter, validation, etc.)
- `test/` - Integration tests
- `todo/` - Project task tracking (YAML files)
- `LLM-GUIDE.md` - Quick reference for LLMs
