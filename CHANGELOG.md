# Changelog

All notable changes to llm-todo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.8.2] - 2026-01-26

### Fixed
- **Typo checker false positive**: Fixed `must_not_do:` field incorrectly flagged as typo
  - Bug: Substring matching caused `must_not:` check to match valid `must_not_do:` fields
  - Impact: Valid YAML files with `must_not_do:` instructions were rejected
  - Fix: Line-by-line checking that skips lines containing valid `must_not_do:` before checking for `must_not:` typo
  - Typo detection still works correctly for actual `must_not:` typos

- **CRITICAL: UpdateTasksFromYAML cross-session contamination**
  - Bug: `UpdateTask()` didn't validate session ownership, allowing enrichment imports to update wrong session's tasks
  - Impact: Importing enrichment file in project A could silently update tasks in project B (same database, different sessions)
  - Example: Importing `vibe-session.yaml` with task IDs 1-20 would update tasks 1-20 from ANY session
  - Fix: Added `UpdateTaskInSession()` that validates `session_id` in WHERE clause
  - Fix: `UpdateTask()` now checks `RowsAffected` and returns error if task doesn't exist
  - Enrichment imports now correctly fail with clear error if tasks don't exist in target session

## [0.8.1] - 2026-01-24

### Fixed
- **CRITICAL: BulkCreateTasks ID calculation bug** causing dependency updates to silently fail
  - Bug: Assumed SQLite's `LastInsertId()` returns FIRST inserted ID, but it returns LAST
  - Impact: Multi-row INSERTs calculated wrong IDs ([5,6,7] instead of [3,4,5])
  - Result: Dependency UPDATE statements targeted non-existent task IDs (0 rows affected)
  - Fix: Calculate IDs backward from lastID: `lastID - (numTasks - 1 - i)`
  - Caught by integration tests before production deployment
- Dependencies now display correctly in `llmtodo show <task-id>` output
- Removed emoji from blocked task output (LLM optimization: `⚠️ Blocked` → `BLOCKED`)
- Fixed enrichment workflow test expectation (all tasks updated when in YAML, not just enriched ones)

### Benefits
- Task dependencies now work correctly for bulk imports
- Integration test suite prevented critical data corruption bug from reaching production
- All 15 integration tests passing

## [0.8.0] - 2026-01-24

### Added
- **LLM-Optimized Enrichment Workflow**: Auto-generated enrichment files for adding context to tasks
  - Template-based approach: ONE example task showing ALL available fields
  - Minimal actual tasks: just `id` and `title` (add fields as you know them)
  - Clear instruction: "DO NOT EDIT IN PLACE - Re-write ENTIRE output and overwrite this file"
  - File location: `~/.llm-todo/enrichment/{session-id}.yaml`
  - Terminal output shows LLM workflow: Read → Write (overwrite) → Import
  - Update-by-ID support: Re-importing enrichment file updates existing tasks (no duplicates)
  - Token efficient: ~3,000 tokens vs 6,000-12,000 for TodoWrite (50-75% savings)
  - Persistent: File survives session end, reviewable by humans

- **Import validation**: `llmtodo import --validate <file.yaml>` checks YAML validity without importing
  - Parse-only validation (no database access)
  - Catches typos, invalid fields, YAML syntax errors
  - Shows what would be imported

- **Atomic bulk imports**: Refactored import to use multi-row INSERT
  - All tasks imported in single SQL command (all-or-nothing)
  - Faster: fewer database round trips
  - Better for database performance
  - Prevents partial imports on error

- **Delete command**: Permanently delete tasks from database
  - `llmtodo delete <task-ids>` - delete by ID
  - Aliases: `rm`, `remove`
  - Batch deletion: `llmtodo delete 1,2,3`
  - Cannot be undone (use `done` to keep in database)

### Fixed
- **Blocking interactive prompts removed**: `code` and `research` commands no longer prompt by default
  - OLD: Prompted for goal, boundaries, success criteria (blocking workflow)
  - NEW: Quick creation by default, optional `--prompt` flag for interactive mode
  - Changed `--skip-prompt` to `--prompt` (opt-in instead of opt-out)
  - Enrichment files provide better workflow for adding context after creation

- **Import error messages dramatically improved**:
  - Typo detection: catches "priortiy" → suggests "priority"
  - Field validation: invalid priority/effort/type with suggestions
  - Task ID validation: shows available tasks when ID not found
  - YAML syntax errors: suggests checking indentation and shows template command
  - Clear text prefixes: ERROR, HINT, WARNING (LLM-optimized, no emojis)

### Changed
- **Enrichment file format redesigned for LLM one-shot generation**:
  - Before: Every task had all fields with empty values and inline comments (repetitive, 115 lines for 12 tasks)
  - After: ONE example task + minimal actual tasks (45 lines for 12 tasks, ~60% reduction)
  - LLMs can now overwrite entire file in one Write tool call instead of 15+ Edit tool calls
  - Humans can still edit files normally in vim/editor

- **LLM-optimized output**: Removed emojis from error messages, replaced with text prefixes
  - Before: ❌, 💡, ⚠️ (visual aids for humans, noise for LLMs)
  - After: ERROR, HINT, WARNING (clear semantic meaning, fewer tokens)
  - Whitespace optimization: every character counts in LLM token budgets

### Benefits
- **Mid-stream enrichment**: When LLM has context from working together, can enrich all tasks in one shot
- **Natural LLM workflow**: Read template once, generate complete enriched version, overwrite file, import
- **No prompting friction**: Create tasks fast, enrich later via file (or skip entirely)
- **Persistent task context**: Unlike TodoWrite, enrichment files survive session end
- **Human-friendly too**: File-based workflow works great for both LLMs and humans

## [0.7.0] - 2026-01-23

### Added
- **Auto-switch on import**: Import command now automatically switches to the imported session for immediate context
- **Priority breakdown display**: Import shows task distribution by priority (p0: 21 tasks, p1: 13 tasks, etc.)
- **GetPriorityStats() function**: New internal function to query task counts by priority level
- **Actionable next steps**: Import output now shows specific commands to run next (`llmtodo next`, `llmtodo get p0`)

### Fixed
- **Project-local database detection**: Now detects project directories by presence of `.git` or `.llm-todo` directory
  - OLD: Only used project-local DB if `.llm-todo/tasks.db` file already existed
  - NEW: Automatically creates project-local DB when copying binary to any git repository
- **Project-local session state**: Current session tracking now respects project boundaries
  - Session state stored in `.llm-todo/current` for project-local, `~/.llm-todo/current` for global

### Changed
- **Import UX dramatically improved for LLM orientation**:
  - Before: Import succeeded but required manual discovery (`sessions`, `use`, `status`)
  - After: One command imports, switches, and shows priority breakdown with actionable next steps
  - Token efficiency: Eliminates 3-4 follow-up commands for LLM to understand what was imported

### Benefits
- **Zero-friction project setup**: Copy binary to new repo, run import, and start working - no manual session switching
- **LLM clarity**: Import output answers "What changed?", "Where am I?", "What should I do next?"
- **Consistent behavior**: Project-local detection works the same way for both database and session state

## [0.6.0] - 2026-01-22

### Added
- **Session Naming System** for managing multiple work streams in one directory:
  - `--name` flag on `quick`, `code`, `research` commands creates `{directory}-{name}` sessions
  - Auto-generated timestamp fallback: `{directory}-{timestamp}` (YYYYMMDD-HHMM format)
  - Naming hint displayed after auto-generated sessions to encourage clarity
  - Examples: `llmtodo quick "task1" --name feature-x` or auto `llm-todo-20260122-1430`
- **Session Archiving** for reducing clutter in active sessions list:
  - `llmtodo session archive <session-id>` to hide completed sessions
  - `llmtodo session restore <session-id>` to bring back archived sessions
  - `llmtodo sessions --archived` to view archived sessions separately
  - Archived sessions excluded from default `sessions` list
  - Sessions table schema updated to support `archived` status
- **Session Scope Filtering** via `--session` flag on all query commands:
  - Works with: `get`, `show`, `status`, `next`, `done`, `block`, `note`, `search`
  - Query different sessions without switching: `llmtodo get p0 --session other-project`
  - Auto-sets current session when creating tasks with `quick`/`code`/`research`
- **LLM Quick Reference Guide**:
  - `llmtodo guide` command shows token-efficient workflows
  - Covers: session creation, work loop, batch operations, context switching
  - Zero tokens until invoked (opt-in reference)
  - Replaces need for verbose --help documentation
- **Smart Context Warnings** in `next` output:
  - Detects missing files, instructions, notes, and refs
  - Shows "MISSING CONTEXT" warning with enrichment hint
  - Inline display of enrichment when present (instructions, files, refs)
  - Saves 500+ tokens by preventing unnecessary `show` calls

### Changed
- **Improved `next` output header**:
  - Old: `NEXT: Task Title (#123)` (confusing - is #123 next or current?)
  - New: `Task #123: Task Title` (clear - this IS the task)
  - Removes cognitive overhead about command vs output naming
- **Auto-set current session** on task creation:
  - `quick`, `code`, `research` now automatically call `SetCurrentSession()`
  - Ensures `llmtodo list` shows newly created tasks by default
  - Solves problem of tasks being created but not visible in current session
- **Session command arg handling**:
  - Increased max args from 2 to 3 to support `session goal <id> "<text>"`
  - Fixed goal suggestion hint to remove unnecessary quotes around session ID

### Fixed
- Session goal suggestion now shows correct syntax without extra quotes
- Query commands now properly respect current session instead of showing all sessions

### Benefits
- **Multiple work streams**: Create separate sessions for different features in same directory
- **Clean workspace**: Archive completed sessions while preserving history
- **Cross-session queries**: Check other sessions without context switching
- **Token efficiency**: Context warnings prevent wasteful `show` calls (88% savings)
- **Better UX**: Auto-session-setting means tasks are immediately visible after creation

## [0.5.0] - 2026-01-19

### Added
- **Task Enrichment System** for cold-start LLM sessions:
  - 5-type enrichment scoring (0-5 scale): instructions, files, output, context, dependencies
  - `llmtodo enrich <task-id>` command with non-blocking flag-based interface
  - `--must-do` and `--must-not` flags for actionable instructions
  - `--files` for related file paths
  - `--output` for concrete deliverables
  - `--notes` for WHY context (with `--replace-notes` flag)
  - `--deps` for prerequisite task dependencies
  - `llmtodo enrich <id> --status` shows enrichment completeness (X/5)
  - `llmtodo enrich <id> --suggest` provides specific enrichment examples
  - Session enrichment via `llmtodo enrich --session` with `--goal`, `--boundaries`, `--success` flags
- **LLM-Friendly Guidelines** embedded in enrichment hints:
  - Concise > verbose (no prose)
  - Imperative > descriptive
  - Factual > emotional
  - No emojis, no past-tense storytelling
  - Context answers WHY, not WHAT was built

### Changed
- **Cleaner output for LLM consumption**:
  - Removed emojis from enrichment hints and suggestions
  - Changed "Tip:" to "RECOMMEND:" (directive instead of suggestion)
  - Simplified task creation output (removed checkmarks, made commands explicit)
  - Session goal suggestion now concise and factual
- **Enrichment hint examples** now show style guidelines:
  - Instructions: "Concise action item" (not generic "item1")
  - Context: "Why this task exists" (not generic "context")
  - Output: "Concrete deliverable" (not generic "description")

### Benefits
- Stateless LLMs can orient on tasks without conversation history
- Task enrichment persists across sessions (unlike TodoWrite)
- Enrichment hints teach users how to write LLM-friendly context
- 95%+ token savings vs TodoWrite while maintaining context quality
- Session + task enrichment provides two-tier orientation (project-level + task-level)

## [0.4.0] - 2026-01-19

### Added
- **Enhanced `llmtodo session` command** for comprehensive context discovery:
  - Default view shows recent 5 completed, next 5 pending (p0), and all blocked tasks
  - `--all` flag shows complete task lists grouped by status
  - Progress percentage display (e.g., "66/99 completed (66%)")
  - Blocking reasons displayed inline for context
  - Solves fresh LLM cold-start problem: one command provides full project state
- **Comprehensive help documentation**:
  - `llmtodo get --help` now shows detailed usage for priority/status filters
  - `llmtodo show --help` explains what information is displayed
  - `llmtodo session --help` documents context discovery features
  - Examples for all flags (--all, --full)
- **Session goal visibility**:
  - `llmtodo sessions` table now displays goal column (was computed but hidden)
  - Session goal appears in `next` fallback with actionable discovery hint

### Changed
- **Token optimization for LLM readability**:
  - Removed emojis from high-frequency outputs (NEXT, SUGGESTIONS, NOTES, etc.)
  - Changed visual symbols to text: `✓` → `[done]`, `⚠️` → `[blocked]`, `•` → `-`
  - Kept emojis only in low-frequency human feedback (success messages)
- **Improved session context in `next` output**:
  - Old: `📋 Session: llm-todo (code)` / `🎯 Goal: ...` (vague, no action)
  - New: `Session: llm-todo - <goal>` / `Session details: llmtodo session` (actionable)
  - Removed unhelpful session type "(code)" from context display
  - Added discovery hint teaching LLMs how to get full context
- **Session management help** moved to dedicated section in main help output

### Performance
- Token paradox solved: investing ~14 tokens in discovery hints saves ~186 tokens in exploration
- LLM can now orient itself in fresh session with 1-2 commands instead of 4-6

### Benefits
- Fresh LLM instances can understand project state in <3 commands:
  1. `llmtodo next` - see current task + session goal + discovery hint
  2. `llmtodo session` - see recent work, next priorities, blockers
  3. Ready to work with full context
- Cleaner, more scannable output for LLM token parsers
- Command discoverability through inline hints

## [0.3.0] - 2026-01-19

### Added
- **Session Goal/Context Feature** for cold-start LLM sessions:
  - `--goal` flag for `quick`, `code`, and `research` commands
  - `llmtodo session` command to view current session context
  - `llmtodo session goal "<text>"` to set/update session goal
  - Session goal displayed in `next` output when available
  - YAML import now supports top-level `goal:` field
  - Educational post-creation message when goal is missing:
    - Explains WHY session context matters (cold-start sessions)
    - Shows exact command to add context
    - Includes example goal text
- Non-interactive flags for LLM usage:
  - `--skip-prompt` flag for `code` and `research` commands
  - `--boundaries` and `--success` flags for `code` command
  - `--deliverables` flag for `research` command

### Changed
- `quick/code/research` commands now suggest adding session goal if not provided
- `import` command checks for goal in YAML and updates session automatically
- `next` output shows session context as fallback when task instructions are sparse

### Benefits
- Solves cold-start problem: LLMs can understand project context across sessions
- Reduces need for verbose task instructions when session goal provides umbrella context
- Token-efficient: session goal shared across all tasks vs repeating context per-task

## [0.2.0] - 2026-01-19

### Added
- **Maximum LLM Clarity** output format for `get <priority>` commands:
  - ACTIVE/QUEUED/BLOCKED grouping for instant context scanning
  - Smart 10-item limit on QUEUED section (saves ~700 tokens per call)
  - "showing X/Y" hints when lists are truncated
  - `--full` flag to show complete lists without limits
  - `--all` flag to include COMPLETED section (hidden by default)
- Priority filters now exclude completed tasks by default (focus on actionable work)

### Changed
- `get p0`, `get p1`, etc. now show only pending + in_progress tasks by default
  - Old behavior: showed all tasks including completed (noisy, token-heavy)
  - New behavior: completed hidden unless `--all` specified
- Output format optimized for LLM parsing and token efficiency
  - Grouped sections with clear headers (ACTIVE, QUEUED, BLOCKED, COMPLETED)
  - Summary line shows counts at a glance: "p0: 2 active, 5 queued"

### Performance
- Token savings: ~700 tokens per `get p0` call by excluding completed tasks
- Reduced cognitive load for LLMs with clearer, structured output

## [0.1.0] - 2026-01-19

### Added
- Initial release of llmtodo
- Core task management with three session modes:
  - `quick` - Fast 3-5 task sessions
  - `code` - Structured development projects (20+ tasks)
  - `research` - Non-code investigation workflows
- Template system with 3 built-in vibe workflow templates:
  - `vibe-crud-db` - Full CRUD with database (17 tasks)
  - `vibe-crud-no-db` - Simple endpoints without database (10 tasks)
  - `vibe-update-endpoint` - Modify existing endpoints (11 tasks)
- 12 command aliases for LLM discoverability:
  - Natural language: `add`, `create`, `list`, `ls`
  - MCP-style: `fetch`, `find`, `read`, `view`, `extract`, `analyze`
- Priority system (p0-p4) with sub-ordering
- Batch operations: `llmtodo done 1,2,3`
- Cross-session persistence with SQLite (WAL mode)
- Conditional output - only shows fields that are set
- File tracking and search capabilities
- YAML import for bulk task creation
- Token efficiency: 95-99% savings vs TodoWrite (empirically proven)

### Documentation
- Comprehensive README with installation instructions
- LLM-GUIDE.md for quick reference
- Empirical token efficiency testing results
- Integration test suite (9 tests)

[Unreleased]: https://github.com/dtnitsch/llm-todo/compare/v0.8.1...HEAD
[0.8.1]: https://github.com/dtnitsch/llm-todo/compare/v0.8.0...v0.8.1
[0.8.0]: https://github.com/dtnitsch/llm-todo/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/dtnitsch/llm-todo/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/dtnitsch/llm-todo/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/dtnitsch/llm-todo/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/dtnitsch/llm-todo/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/dtnitsch/llm-todo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/dtnitsch/llm-todo/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dtnitsch/llm-todo/releases/tag/v0.1.0
