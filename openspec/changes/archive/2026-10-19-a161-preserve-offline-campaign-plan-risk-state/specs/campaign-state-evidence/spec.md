## ADDED Requirements

### Requirement: Offline capture preserves native campaign plan and risk evidence
The system SHALL provide one-way offline capture of KR and US continuation, reversal and weekly plan and risk state into detached inspection evidence. It SHALL preserve complete immutable plan fields and native digest, original initial budget, currency and frozen FX data, recorded filled and held risk, all native latch values and native deduplication records without changing source state.

#### Scenario: Nonzero risk survives reopen
- **WHEN** a complete plan and nonzero native risk snapshot are encoded, saved to a new temporary file, closed and reopened
- **THEN** inspection restores all captured plan and risk values and native digests without creating a new zero risk state or native execution object

#### Scenario: Source state changes after capture
- **WHEN** caller-owned maps or FX values are changed after capture completes
- **THEN** previously captured evidence remains unchanged and detached from the source

### Requirement: Recorded risk retains family-specific uncertainty
Inspection SHALL distinguish initial budget from recorded filled and held values, preserve unknown and overbudget states and retain each family's original fill/cancel semantics. It SHALL mark weekly cancel data as absent from RiskState and separate ReservationState as not captured. It MUST NOT interpret weekly fill Known as verified actual risk, invent missing cancel records or repair unknown amounts to zero.

#### Scenario: Unknown and excess risk remain visible
- **WHEN** a valid captured state contains an unknown-risk latch, incomplete native fill result or amounts exceeding its initial budget
- **THEN** restoration preserves the recorded values and uncertainty without clearing latches or clamping values

#### Scenario: Weekly identity is known but actual risk is unknown
- **WHEN** a weekly fill has Known true and its risk state has unknown-actual-risk latched
- **THEN** the identity flag remains true while actual risk remains unverified and the missing separate reservation history is explicit

### Requirement: Evidence restoration is bounded and externally scoped
The decoder SHALL accept only canonical schema-v1 evidence of at most 4 MiB matching an externally supplied SHA-256 content digest and expected account, market, symbol, campaign, lane, lane version and generation. It SHALL reject incomplete, malformed, duplicate-key, unknown-field, trailing, unsupported-version, oversized, corrupted and misbound input without a usable partial result.

#### Scenario: Tampered or foreign evidence
- **WHEN** any captured bytes or expected scope differ, including a self-rehashed altered artifact with the original expected digest
- **THEN** restoration fails and produces no usable evidence

#### Scenario: Missing or malformed evidence
- **WHEN** evidence omits a required value or uses duplicate keys, an unsupported version, trailing data or exceeds the size limit
- **THEN** restoration fails without inferring legacy facts or zero risk

### Requirement: Restored evidence cannot confer execution authority
Every restored result SHALL identify itself as inspection-only caller-provided data with capture time and unverified ledger currentness. Export and decode SHALL NOT mint native plan, FX or RiskState authority, refresh FX, calculate qFinal, reserve risk, create a lease, alter the journal or connect to production execution. Snapshot integrity MUST NOT be described as provenance authentication or real-account completeness.

#### Scenario: Snapshot is older than live activity
- **WHEN** a valid stored snapshot is inspected after subsequent real activity may have occurred
- **THEN** only its recorded capture-time values are returned with unverified currentness and no order permission

#### Scenario: Existing journal lacks a full plan
- **WHEN** only a160 journal rows or digests exist without captured plan and risk evidence
- **THEN** the existing unavailable result remains unchanged and no full plan or actual risk is synthesized
