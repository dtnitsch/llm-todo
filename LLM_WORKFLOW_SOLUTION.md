# LLM-Optimized Enrichment Workflow

## The Problem We Solved

**Original workflow assumption:** LLMs would edit enrichment files incrementally (Edit tool x15 for 15 tasks)

**Reality:** LLMs generate structured data in one shot, not sequential edits

**Token cost comparison:**
- Sequential edits: 18+ tool calls for 15 tasks
- TodoWrite: 1 tool call, 6,000-12,000 tokens (not persistent)
- **New workflow: 3 tool calls, ~3,000 tokens (persistent + reviewable)**

## The Solution

### One-Shot Overwrite Workflow

**For LLMs:**
1. Read enrichment file ONCE (template + minimal tasks)
2. Write tool ONCE (overwrite entire file with enriched version)
3. Import updated file

**For Humans:**
1. Edit enrichment file in vim/editor
2. Import updated file

### Example-Driven Template

**Generated enrichment file structure:**

```yaml
# Auto-generated enrichment file for session: llm-todo-20260124-1500
# LLM: DO NOT EDIT IN PLACE - Re-write ENTIRE output and overwrite this file
# Re-import after re-writing: llmtodo import ~/.llm-todo/enrichment/session.yaml

goal: ""  # Optional: describe the purpose of this session

tasks:
  # EXAMPLE TASK - Shows all available fields - REMOVE THIS FROM YOUR OUTPUT
  - id: example-task
    title: "Example: Implement feature X"
    priority: p0              # p0 (critical), p1 (important), p2 (normal)
    effort: m                 # xs, s, m (effort estimate)
    type: task                # task, research, coordination, analysis
    files:
      - "path/to/file.go"
    refs:                     # Reference URLs or docs
      - "https://docs.example.com"
    instructions:
      must_do:
        - "Add validation"
        - "Handle edge cases"
      must_not_do:
        - "Don't break existing API"

  # Your actual tasks (add fields as you know them)
  - id: task-55
    title: "Implement auth"

  - id: task-56
    title: "Add tests"

  - id: task-57
    title: "Deploy"
```

**Key features:**
- ✅ Clear instruction: "DO NOT EDIT IN PLACE"
- ✅ One example showing ALL available fields
- ✅ Actual tasks are MINIMAL (just id and title)
- ✅ LLM fills in what they know (priority, effort, files, etc.)

## Mid-Stream Use Case

**Scenario:** After working together for 30 minutes, LLM recommends next steps

**LLM has context:**
- ✅ Knows file structure (been working in codebase)
- ✅ Knows priorities (p0 critical bugs vs p2 nice-to-haves)
- ✅ Knows effort (simple fix vs complex refactor)
- ✅ Knows constraints (from conversation: "don't break API")

**Workflow:**

```bash
# 1. Create 15 tasks
./ltd code "Implement auth" "Add tests" ... "Deploy"

# 2. Read enrichment file (see template)
Read tool → ~/.llm-todo/enrichment/session.yaml

# 3. Overwrite with enriched version (ONE Write call)
Write tool → Overwrite entire file:

goal: "Implement OAuth2 auth system"

tasks:
  - id: task-64
    title: "Implement auth"
    priority: p0
    effort: m
    files: ["internal/auth/manager.go", "internal/auth/jwt.go"]
    instructions:
      must_do: ["Add token validation", "Handle refresh"]
      must_not_do: ["Don't break existing session API"]

  - id: task-65
    title: "Add tests"
    priority: p0
    effort: s
    files: ["internal/auth/manager_test.go"]

  # ... 13 more tasks

# 4. Import
./ltd import ~/.llm-todo/enrichment/session.yaml
```

**Result:**
- ✅ 15 tasks with rich context
- ✅ Persistent (survives session end)
- ✅ Reviewable (file on disk)
- ✅ Token efficient (~3,000 tokens vs 10,000 for TodoWrite)

## Benefits Over TodoWrite

| Feature | TodoWrite | Enrichment File |
|---------|-----------|-----------------|
| Tool calls | 1 | 3 |
| Tokens | 6,000-12,000 | ~3,000 |
| Persistent | ❌ Lost on session end | ✅ Saved to disk |
| Reviewable | ❌ Can't inspect later | ✅ File on disk |
| Editable | ❌ Regenerate from scratch | ✅ Edit and re-import |
| Human-friendly | ❌ JSON in API | ✅ YAML file |

## Terminal Output

**After creating tasks:**

```
Created 15 tasks:
  64. Implement auth
  65. Add tests
  ...

Pre-filled enrichment: ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

RECOMMEND: Enrich tasks for better context
  LLM workflow:
    1. Read: ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml
    2. Write: Overwrite entire file with enriched version
    3. Import: llmtodo import ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

  Human workflow:
    1. Edit: vim ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml
    2. Import: llmtodo import ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

Skip enrichment: llmtodo next
```

## When LLMs Will Use This

**Cold start:** NO
- Don't have context yet
- Would have to explore first, then enrich, then re-import
- Better to just start working and clarify conversationally

**Mid-stream:** YES
- Have context from working together
- Know files, priorities, effort, constraints
- Want persistence for multi-session projects
- Token efficient (50-75% savings vs TodoWrite)

## Implementation

### Files Changed
- `internal/exporter/enrichment.go` - Generate LLM-optimized template
- `cmd/llmtodo/mode_commands.go` - Show LLM workflow in output
- `internal/exporter/enrichment_test.go` - Test new format

### Files Created
- `LLM_WORKFLOW_SOLUTION.md` - This document
- `/tmp/sample-enrichment-new.yaml` - Example output

## The Verdict

**The enrichment file workflow is now optimized for both LLMs and humans.**

**LLMs will use it mid-stream when they have context and want persistence.**

**Humans will use it anytime for file-based task management.**

**Both workflows are fast, efficient, and natural for their respective users.**
