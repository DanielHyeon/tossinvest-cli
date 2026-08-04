## ADDED Requirements

### Requirement: unattended attestation renewal uses one operating profile

The unattended renewal command SHALL use the same explicit config profile as the operator console's
read-only survey. The soak record, credentials, attestation output and engine-interlock input SHALL be
resolved through that profile's normal path rules and SHALL NOT mix an implicit data-profile record with
a config-profile output. Existing evidence SHALL NOT be copied between profiles to satisfy this
requirement.

#### Scenario: renewal consumes the console survey record

- **WHEN** the console-profile survey has produced qualifying evidence and unattended renewal runs
- **THEN** renewal reads that record and writes the attestation where the same profile's engine interlock reads it

#### Scenario: legacy data-profile evidence remains separate

- **WHEN** a stale legacy record exists under the data profile
- **THEN** renewal for the console profile does not read, move or relabel that record

### Requirement: renewal failure is visible before expiry blocks startup

A refused or failed unattended renewal SHALL preserve a failed status and its reasons. The normal
operator surface SHALL report renewal failure or impending attestation expiry before the active
attestation expires, without requiring an engine restart to reveal the problem. The warning SHALL NOT
stop a running engine and SHALL NOT weaken or bypass the startup interlock.

#### Scenario: incomplete evidence is reported before expiry

- **WHEN** renewal cannot issue a fresh attestation while the active attestation approaches expiry
- **THEN** the operator sees the failed status, unmet criteria and expiry horizon before startup is affected

#### Scenario: a running engine is not stopped by the warning

- **WHEN** the active attestation enters the warning window while an engine is already running
- **THEN** the system reports the condition without stopping that engine or changing the next-start interlock decision

### Requirement: attestation renewal preserves operational safety state

Preparing, installing and verifying unattended renewal SHALL NOT change trading, Guardian, lane,
kill-switch, position-adoption or automation-gate settings. Installing or activating an external service
definition SHALL require explicit human approval. Survey execution SHALL remain read-only and no live
order command SHALL be invoked as part of renewal verification.

#### Scenario: renewal is deployed without changing trading state

- **WHEN** the aligned renewal service is installed and verified with human approval
- **THEN** all operating toggles and risk settings remain byte-for-byte unchanged and no order-side effect occurs
