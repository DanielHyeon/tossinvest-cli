## Context

The repository already has two distinct concerns around agent saves:

- `.codex/hooks.json` sends file-edit observations to `tools/sdd-history/agent_save_hook.py`.
- The operator's Claude configuration invokes the untracked root `save-session.sh`, which reads Claude-specific state and writes `.ai-context/session-summary.md`.

Official OpenAI hook documentation defines project-local `hooks.json`, `PostToolUse` matching for `Bash` and `apply_patch`, one JSON object on stdin, and common fields including `session_id`, `transcript_path`, `cwd`, and `model`. It also states that transcript internals are not a stable hook interface. Matching command hooks can run concurrently and require explicit trust.

This is repository tooling only. It does not touch trading, account state, authentication, orders, or runtime services.

## Goals / Non-Goals

**Goals:**

- Preserve a current, human-readable Codex handoff after Bash and file-edit tool calls.
- Keep Codex generated context physically and logically separate from Claude generated context.
- Preserve the existing SDD observation hook and all Claude files unchanged.
- Avoid delaying the agent loop and avoid corrupt output under concurrent hook calls.
- Fail open when hook input, Git metadata, or the optional transcript is unavailable.

**Non-Goals:**

- Changing Codex's own transcript storage location or transcript schema.
- Sharing memory between Claude and Codex.
- Treating the generated summary as an authoritative project record.
- Installing third-party packages or uploading session data.

## Decisions

### 1. Use a separate Codex implementation and store

Add `.codex/hooks/save_session.py` and write only below `<repo>/.codex-context/`. Do not call, import, or parameterize the Claude `save-session.sh`, and do not access `.claude/` or `.ai-context/`.

Alternative considered: add flags to `save-session.sh`. Rejected because it makes Claude's currently untracked script a shared dependency and creates a regression surface the operator explicitly prohibited.

### 2. Use the documented hook envelope and defensive transcript parsing

The saver consumes the JSON envelope from stdin and uses documented fields for session metadata. If `transcript_path` names a readable local file, it best-effort extracts recent `user` and `assistant` message text from currently observed `response_item/message` records. Unknown records and malformed lines are ignored. The summary never copies tool inputs, tool outputs, reasoning records, developer messages, or the transcript path.

Alternative considered: copy the transcript wholesale. Rejected because it would retain excessive data, tool output, and potentially sensitive content while depending more strongly on an unstable format.

### 3. Bound and sanitize persisted conversation context

Retain at most 30 recent user/assistant message entries, truncate each entry, and redact common credential assignments and token formats before writing. Git status, staged/unstaged statistics, recent commits, and recently modified repository files provide the stable handoff context.

### 4. Serialize writes and replace atomically

Use a non-blocking process lock under `.codex-context/`. A concurrent invocation that cannot acquire the lock exits successfully. Write a temporary file in the destination directory, `fsync`, then atomically replace `session-summary.md`. Before replacement, create a uniquely named backup and keep the newest five.

Alternative considered: allow all asynchronous hooks to write. Rejected because Codex can launch matching hooks concurrently, causing backup collisions or partial summaries.

### 5. Run asynchronously and fail open

Configure the new handler with `async: true` so repository scans do not delay tool-result processing. The script catches operational failures, emits only a concise warning to stderr, and returns zero. The generated summary is advisory and must not block Codex.

### 6. Test the boundary, not Codex internals

Unit tests invoke the saver against temporary project/output directories and synthetic hook/transcript JSON. They verify isolated paths, extraction/redaction, malformed input, backup retention, and hook configuration. The tests do not modify or depend on the user's actual Codex or Claude session stores.

## Risks / Trade-offs

- **Transcript records change shape** → Git and hook metadata still save; unknown transcript records are skipped without failing the hook.
- **A secret uses an unrecognized format** → Only bounded user/assistant text is retained in a local ignored directory; tool output and reasoning are never retained. The summary remains advisory and should not be shared.
- **Many tool calls launch many asynchronous hooks** → A non-blocking lock permits one writer and drops redundant overlapping saves.
- **New hook is installed but not active** → Codex requires the operator to approve the exact hook hash through `/hooks`; installation documentation and final handoff state this explicitly.
- **Generated files dirty Git** → `.codex-context/` is root-ignored and tests assert the ignore rule.

## Migration Plan

1. Add the contract and tests.
2. Add the Codex-only saver and `.codex-context/` ignore rule.
3. Append the new matcher group to `.codex/hooks.json` without replacing the existing group.
4. Run the saver with a synthetic PostToolUse event and verify `.codex-context/session-summary.md`.
5. The operator opens `/hooks` and trusts the new definition. Rollback removes only the new Codex matcher, script, tests, and `.codex-context/` ignore rule; Claude files remain untouched.

## Open Questions

None. The store boundary and event coverage were explicitly selected by the operator.
