# Changelog

All notable changes to llm-todo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/dtnitsch/llm-todo/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dtnitsch/llm-todo/releases/tag/v0.1.0
