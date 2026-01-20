# Changelog

All notable changes to llm-todo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/dtnitsch/llm-todo/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/dtnitsch/llm-todo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/dtnitsch/llm-todo/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dtnitsch/llm-todo/releases/tag/v0.1.0
