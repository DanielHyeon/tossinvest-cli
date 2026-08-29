## Why

Codex currently records the generic SDD save event after edits, but it does not preserve a human-readable Codex session handoff after Bash and file-edit tools. Reusing the Claude script would mix Claude state and `.ai-context` data into Codex output, so Codex needs an isolated saver and store.

## What Changes

- Add a Codex-only `PostToolUse` handler for Bash and `apply_patch` while preserving the existing SDD observation handler.
- Add a Codex-only session saver that consumes the documented hook JSON input and defensively reads the referenced Codex transcript.
- Store generated summaries and bounded backups only under `.codex-context/`, independently from Claude's `.ai-context/` store.
- Ignore generated Codex context in Git and add automated tests for isolation, malformed input, backup retention, atomic output, and hook wiring.
- Leave `.claude/settings.json` and `save-session.sh` unchanged.

## Capabilities

### New Capabilities

- `codex-session-save`: Codex lifecycle-hook session capture, local storage isolation, bounded backups, and failure-safe operation.

### Modified Capabilities

None.

## Impact

- Affected Codex configuration: `.codex/hooks.json`
- New Codex-only implementation and tests: `.codex/hooks/save_session.py`, `tools/sdd-history/test_codex_session_save.py`
- New ignored generated-data root: `.codex-context/`
- No trading runtime, account state, live-order path, external dependency, Claude hook, or Claude session script is changed.
