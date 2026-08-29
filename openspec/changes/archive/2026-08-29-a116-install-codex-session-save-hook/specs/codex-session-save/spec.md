## ADDED Requirements

### Requirement: Codex PostToolUse session capture
The repository SHALL configure a Codex `PostToolUse` command hook for `Bash` and `apply_patch`. The new handler SHALL coexist with the existing SDD agent-save handler and SHALL execute asynchronously without changing normal tool results.

#### Scenario: Codex completes a shell command
- **WHEN** Codex completes a tool call whose hook name is `Bash`
- **THEN** the Codex session saver is scheduled without replacing or disabling the SDD agent-save handler

#### Scenario: Codex applies a file patch
- **WHEN** Codex completes a tool call whose hook name is `apply_patch`
- **THEN** the same Codex session saver is scheduled

### Requirement: Agent-specific storage isolation
The Codex session saver SHALL write generated summaries and backups only below the repository-local `.codex-context/` root. It SHALL NOT read from or write to `.claude/`, `.ai-context/`, or the Claude `save-session.sh`, and `.codex-context/` SHALL be excluded from Git.

#### Scenario: Codex summary is saved
- **WHEN** the Codex session saver handles a valid hook event
- **THEN** the latest summary is `.codex-context/session-summary.md` and no Claude-owned path is modified

### Requirement: Bounded useful handoff
The latest summary SHALL include timestamp, session and model metadata when available, Git working state, recent commits, recent repository files, and bounded recent Codex user/assistant text when a readable transcript is supplied. It SHALL exclude tool payloads, tool responses, reasoning, developer messages, and the transcript filesystem path.

#### Scenario: Transcript contains mixed record types
- **WHEN** the transcript contains user, assistant, developer, reasoning, and tool records
- **THEN** only bounded user and assistant message text is eligible for the generated summary

#### Scenario: Message text contains a recognized credential
- **WHEN** eligible message text contains a recognized credential assignment or token format
- **THEN** the persisted summary replaces the credential value with a redaction marker

### Requirement: Failure-safe and atomic persistence
The saver SHALL return success for malformed hook input, unavailable optional transcripts, Git metadata failures, and concurrent duplicate invocations. A successful writer SHALL publish the latest summary atomically, SHALL keep at most five uniquely named backups, and SHALL never expose a partially written latest summary.

#### Scenario: Hook input is malformed
- **WHEN** stdin is not valid hook JSON
- **THEN** the saver still writes the available repository handoff or exits successfully without affecting Codex tool processing

#### Scenario: Two hook invocations overlap
- **WHEN** one saver holds the destination lock and another invocation starts
- **THEN** the second invocation exits successfully without writing or corrupting the latest summary

#### Scenario: More than five prior summaries exist
- **WHEN** a new latest summary is published
- **THEN** only the five newest backup summaries remain

### Requirement: Claude session saver preservation
The change SHALL NOT modify `.claude/settings.json` or the root `save-session.sh`.

#### Scenario: Codex saver installation is reviewed
- **WHEN** the implementation diff is compared with the captured change base and the pre-existing working tree
- **THEN** no change authored by this change appears in either Claude-owned file
