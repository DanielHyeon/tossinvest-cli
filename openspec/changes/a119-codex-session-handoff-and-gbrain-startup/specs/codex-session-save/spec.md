## MODIFIED Requirements

### Requirement: Codex PostToolUse session capture
The repository SHALL configure a Codex `PostToolUse` command hook for `Bash`,
`apply_patch`, and the additional tool-event names established by sanitized
fixtures from the supported Codex host. The accepted event fixtures and matcher
configuration SHALL be tested together before claiming host coverage. The handler
SHALL coexist with the existing SDD agent-save handler and SHALL execute
asynchronously without changing normal tool results. Session persistence SHALL
continue to satisfy the existing agent-specific storage isolation, bounded
handoff, redaction, and failure-safe atomic persistence requirements.

#### Scenario: Codex completes a shell command
- **WHEN** Codex completes a tool call whose hook name is `Bash`
- **THEN** the Codex session saver is scheduled without replacing or disabling the SDD agent-save handler

#### Scenario: Codex applies a file patch
- **WHEN** Codex completes a tool call whose hook name is `apply_patch`
- **THEN** the same Codex session saver is scheduled

#### Scenario: Supported host emits an additional tool name
- **WHEN** a PostToolUse event uses an additional name established by a sanitized supported-host fixture
- **THEN** the configured matcher schedules the existing isolated saver and the fixture regression verifies the match

#### Scenario: Host coverage has not been observed
- **WHEN** only a configuration or synthetic matcher test has passed and no supported-host event has been observed
- **THEN** runtime event delivery remains unverified and the change cannot claim that an ordinary host tool call refreshed the handoff
