## Why

Codex has a repository-local saver, but its latest handoff remains stale because
the configured PostToolUse matcher does not cover the tool events emitted by the
current Codex host. The same host also attempts to start the TossOS GBrain MCP
twice: one process owns the project PGLite lock while the later attempt exits
with the documented busy status, which Codex surfaces as an incomplete MCP
startup.

## What Changes

- Expand the Codex session-save PostToolUse coverage so a normal Codex tool
  execution schedules the existing isolated saver, while preserving its bounded,
  redacted, atomic `.codex-context/` output.
- Keep Codex session summaries and backups exclusively below
  `.codex-context/`; do not read from or modify Claude-owned files or stores.
- Make the Codex MCP configuration start the TossOS GBrain wrapper exactly once,
  without changing the wrapper's single-writer safety behavior.
- Add regression tests for the accepted Codex event coverage and for one
  Codex-owned GBrain registration.

## Capabilities

### New Capabilities

- `gbrain-codex-mcp-startup`: A Codex workspace starts at most one TossOS GBrain
  MCP process for its project data home.

### Modified Capabilities

- `codex-session-save`: The Codex PostToolUse trigger requirement expands from
  two historical tool names to the events emitted by the current Codex host.

## Impact

- `.codex/hooks.json`, `.codex/hooks/save_session.py`, and
  `tools/sdd-history/test_codex_session_save.py`
- Codex-facing project MCP registration in `.codex/config.toml` and/or
  `.mcp.json`, selected to retain one authoritative Codex launch path
- New and updated OpenSpec specifications and review evidence
- No trading runtime, account state, order path, Claude hook, or
  `save-session.sh` behavior
