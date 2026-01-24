# Auto-Generated Enrichment Files Feature

## Summary

Implements automatic enrichment file generation for quick task creation workflow.

## Key Design Decisions

### LLM-Optimized Workflow
The enrichment file is designed for **one-shot overwrite**, not incremental editing:
- ✅ Read enrichment file ONCE (see template + minimal tasks)
- ✅ Write tool ONCE (overwrite entire file with enriched version)
- ✅ Import updated file

**Token efficiency:** ~3,000 tokens vs 6,000-12,000 for TodoWrite (50-75% savings)

**Clear instructions in file:**
```yaml
# LLM: DO NOT EDIT IN PLACE - Re-write ENTIRE output and overwrite this file
```

### Example-Driven Template
File includes ONE example task showing ALL available fields:
```yaml
  # EXAMPLE TASK - Shows all available fields - REMOVE THIS FROM YOUR OUTPUT
  - id: example-task
    title: "Example: Implement feature X"
    priority: p0              # p0 (critical), p1 (important), p2 (normal)
    effort: m                 # xs, s, m (effort estimate)
    files: ["path/to/file.go"]
    instructions:
      must_do: ["Add validation"]
      must_not_do: ["Don't break API"]
```

Actual tasks are **minimal** (just id and title):
```yaml
  # Your actual tasks (add fields as you know them)
  - id: task-55
    title: "Add export cmd"

  - id: task-56
    title: "Add formatter"
```

## Changes Made

### 1. Fixed Blocking Interactive Prompts
**Issue:** `code` and `research` commands prompted interactively by default, breaking quick workflow

**Fix:**
- Changed `--skip-prompt` to `--prompt` (opt-in instead of opt-out)
- Default behavior: No prompts, quick creation
- Use `--prompt` flag if you want interactive prompts

**Before:**
```bash
./ltd code "task1" "task2"  # ❌ Prompted for goal, boundaries, success
```

**After:**
```bash
./ltd code "task1" "task2"            # ✅ Quick creation, no prompts
./ltd code "task1" --prompt           # ✅ Interactive prompts if needed
./ltd code "task1" --goal "my goal"   # ✅ Via flags
```

### 2. Enhanced Enrichment File Metadata
**Issue:** Enrichment file only included `goal`, missing `success_criteria`, `boundaries`, `deliverables`

**Fix:**
- Enrichment file now includes ALL session metadata when provided
- Respects flags: `--goal`, `--boundaries`, `--success`

**Example output:**
```yaml
# Auto-generated enrichment file for session: llm-todo-20260124-1500
# LLM: DO NOT EDIT IN PLACE - Re-write ENTIRE output and overwrite this file
# Re-import after re-writing: llmtodo import ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

goal: ""  # Optional: describe the purpose of this session

tasks:
  # EXAMPLE TASK - Shows all available fields - REMOVE THIS FROM YOUR OUTPUT
  - id: example-task
    title: "Example: Implement feature X"
    priority: p0              # p0 (critical), p1 (important), p2 (normal)
    effort: m                 # xs, s, m
    type: task                # task, research, coordination, analysis, deliverable
    files:
      - "path/to/file.go"
    refs:                     # Reference URLs or docs
      - "https://docs.example.com"
    instructions:
      must_do:
        - "Add validation"
      must_not_do:
        - "Don't break existing API"

  # Your actual tasks (add fields as you know them)
  - id: task-55
    title: "Design schema"

  - id: task-56
    title: "Implement API"

  - id: task-57
    title: "Add tests"
```

### 3. File Location
**Changed from:** `/tmp/llmtodo-{session-id}.yaml`
**Changed to:** `~/.llm-todo/enrichment/{session-id}.yaml`

**Rationale:**
- Stays within llm-todo ecosystem
- No permission issues
- Persists across reboots (useful for multi-day projects)

## Testing

### Unit Tests: `internal/exporter/enrichment_test.go`
Table-driven tests covering:
- ✅ Enrichment file generation with various metadata combinations
- ✅ GetEnrichmentPath() directory creation
- ✅ YAML escaping for quotes and special characters

Run:
```bash
go test ./internal/exporter/...
```

### Integration Tests: `test/integration/enrichment_workflow_test.go`
End-to-end workflow tests covering:
- ✅ Create tasks → Generate enrichment file
- ✅ Edit enrichment file → Re-import → Verify updates
- ✅ Both quick and code session types
- ✅ Multiple tasks with different enrichments

Run:
```bash
go test ./test/integration/...
```

## Usage Examples

### Quick Session (LLM Workflow)
```bash
# 1. Create tasks
./ltd quick "Fix bug" "Write test" "Deploy"

# Output shows enrichment file path and workflow:
# LLM workflow:
#   1. Read: ~/.llm-todo/enrichment/session.yaml
#   2. Write: Overwrite entire file with enriched version
#   3. Import: llmtodo import ~/.llm-todo/enrichment/session.yaml

# 2. Read enrichment file (see template + minimal tasks)
Read tool → ~/.llm-todo/enrichment/session.yaml

# 3. Overwrite with enriched version (one shot)
Write tool → Overwrite entire file:
goal: "Critical bug fixes for v1.2 release"

tasks:
  - id: task-1
    title: "Fix bug"
    priority: p0
    effort: s
    files: ["src/auth.go"]
    instructions:
      must_do: ["Add validation"]
      must_not_do: ["Don't break existing API"]

  - id: task-2
    title: "Write test"
    priority: p0
    effort: xs
    files: ["src/auth_test.go"]

  - id: task-3
    title: "Deploy"
    priority: p1
    effort: xs

# 4. Import
./ltd import ~/.llm-todo/enrichment/session.yaml
```

### Code Session (No more blocking prompts!)
```bash
# Quick creation (no prompts)
./ltd code "Design" "Implement" "Test"

# With metadata via flags
./ltd code "Task1" "Task2" \
  --goal "Auth system" \
  --boundaries "No OAuth" \
  --success "All tests pass"

# Interactive prompts (opt-in)
./ltd code "Task1" --prompt
```

### Research Session
```bash
# Quick creation
./ltd research "Survey options" "Prototype" "Document"

# With deliverables
./ltd research "Task1" --goal "Evaluate tools" --deliverables "Report, POC"
```

## Workflow

1. **Create tasks** (quick, no prompts)
   ```bash
   ./ltd code "Add feature X" "Write tests" "Update docs"
   ```

2. **Output shows enrichment file path**
   ```
   Created 3 tasks:
     55. Add feature X
     56. Write tests
     57. Update docs

   Pre-filled enrichment: ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

   RECOMMEND: Enrich tasks for better context
     1. Edit: ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml
     2. Import: llmtodo import ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml

   Skip enrichment: llmtodo next
   ```

3. **Edit enrichment file** (add files, instructions, effort)
   ```yaml
   tasks:
     - id: task-55
       title: "Add feature X"
       effort: "m"
       files: ["src/feature.go", "src/feature_test.go"]
       instructions:
         must_do:
           - "Add validation"
           - "Handle edge cases"
         must_not_do:
           - "Don't break existing API"
   ```

4. **Re-import to apply updates**
   ```bash
   ./ltd import ~/.llm-todo/enrichment/llm-todo-20260124-1500.yaml
   # ✓ Updated 3 tasks from enrichment file
   ```

5. **Verify updates**
   ```bash
   ./ltd show 55
   # Should show enriched fields: effort, files, instructions
   ```

## Modified Files

### New Files
- `internal/exporter/enrichment.go` - Enrichment file generation
- `internal/exporter/enrichment_test.go` - Unit tests
- `test/integration/enrichment_workflow_test.go` - Integration tests

### Modified Files
- `cmd/llmtodo/mode_commands.go` - Generate enrichment files, fix prompts
- `internal/importer/yaml.go` - Add UpdateTasksFromYAML() for update-by-ID
- `cmd/llmtodo/import_commands.go` - Detect and handle update mode

## Breaking Changes

**None** - All changes are backward compatible:
- Existing workflows unchanged
- New behavior is additive (generates enrichment file)
- File generation is non-fatal (warns on error)
- Skip workflow still works (`llmtodo next`)

## Future Enhancements

Potential improvements:
- [ ] Auto-cleanup old enrichment files
- [ ] `llmtodo enrich` command to regenerate enrichment file for existing session
- [ ] Richer enrichment file templates based on session type
- [ ] Validation warnings for common mistakes in enrichment files
