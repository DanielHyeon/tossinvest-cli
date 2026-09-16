## ADDED Requirements

### Requirement: One effective Codex project registration
The effective Codex workspace configuration SHALL contain exactly one enabled
registration that starts the TossOS GBrain project wrapper for its project data
home. Selection of the authoritative registration SHALL be based on the
supported host's actual configuration loading behavior and SHALL preserve
registrations owned by other agents. Static configuration tests SHALL distinguish
repository configuration from runtime loading evidence.

#### Scenario: Duplicate Codex registrations target the same project
- **WHEN** the supported Codex host would load two enabled registrations for the same TossOS GBrain project wrapper and data home
- **THEN** the Codex-owned configuration is corrected to one effective registration and a regression rejects the duplicate configuration

#### Scenario: Workspace startup is observed
- **WHEN** the supported Codex host starts the workspace with the reviewed effective configuration
- **THEN** that host launches the project wrapper at most once and the observation is recorded without exposing credentials or session contents

### Requirement: Preserve project ownership and agent isolation
The registration correction SHALL retain the existing project wrapper and its
single-writer lock, busy-exit behavior, and TossOS-specific data home. It SHALL NOT
terminate a lock owner, delete a lock or database, modify Claude-owned hook or
context storage, or change any trading runtime setting.

#### Scenario: Another session owns the project lock
- **WHEN** another session already owns the same project data home
- **THEN** the existing wrapper reports contention through its established busy behavior and the configuration correction does not kill or replace that owner

#### Scenario: Codex configuration repair is reviewed
- **WHEN** the authored configuration and test changes are compared with their captured implementation baseline
- **THEN** Claude-owned hooks and stores, `save-session.sh`, trading controls and account state are unchanged
