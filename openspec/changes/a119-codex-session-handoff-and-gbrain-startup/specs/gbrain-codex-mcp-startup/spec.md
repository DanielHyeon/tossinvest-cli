## ADDED Requirements

### Requirement: One effective Codex project registration
The Codex-effective workspace configuration SHALL contain exactly one enabled
registration that starts the TossOS GBrain project wrapper for its project data
home, and no registration that launches raw `gbrain` (which would bypass the
wrapper's lock). This SHALL be pinned by a static regression against a sanitized
fixture of the registrations the supported host's configuration loader reports.
Static configuration tests SHALL distinguish repository configuration from
runtime loading evidence and SHALL NOT claim that a host started the wrapper.

#### Scenario: A second enabled registration is introduced
- **WHEN** the Codex-owned configuration would enable two registrations for the same TossOS GBrain project wrapper and data home
- **THEN** the regression fails and names both registrations

#### Scenario: A registration launches raw gbrain
- **WHEN** a Codex-owned registration would start `gbrain` directly instead of the project wrapper
- **THEN** the regression fails, because that launch bypasses the single-writer lock

#### Scenario: Concurrent threads start the workspace
- **WHEN** a later Codex thread starts the wrapper while another thread already owns the project lock
- **THEN** the wrapper exits with its documented busy status and the thread reports an incomplete MCP startup, which this change does not remove

### Requirement: Preserve project ownership and agent isolation
This change SHALL retain the existing project wrapper and its single-writer
lock, busy-exit behavior, and TossOS-specific data home. It SHALL NOT terminate a
lock owner, delete a lock or database, modify Claude-owned hook or context
storage, or change any trading runtime setting.

#### Scenario: Another session owns the project lock
- **WHEN** another session already owns the same project data home
- **THEN** the existing wrapper reports contention through its established busy behavior and nothing in this change kills or replaces that owner

#### Scenario: The change is compared with its baseline
- **WHEN** the authored test and fixture changes are compared with the captured implementation baseline
- **THEN** `.codex/hooks.json`, `.codex/config.toml`, `.mcp.json`, `save-session.sh`, Claude-owned hooks and stores, trading controls and account state are unchanged
