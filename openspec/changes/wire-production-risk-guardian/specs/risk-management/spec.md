## ADDED Requirements

### Requirement: Automation gate limits are the production Guardian policy source
The production `RiskGuardian` policy SHALL be derived from the automation gate's
five configured limits and one normalized currency. The Guardian's maximum
quantity, maximum order notional, maximum open exposure, maximum daily loss, and
maximum daily loss ratio SHALL produce a limit snapshot byte-for-byte equivalent
to the interlock's audited snapshot, including configured/set bits and currency.
The per-trade risk budget SHALL equal the configured maximum daily-loss amount in
the same currency.

#### Scenario: USD policy construction
- **WHEN** the gate contains valid USD limits
- **THEN** every Guardian money field and its audited limit snapshot use USD and the configured values without falling back to KRW defaults

#### Scenario: One configured value differs
- **WHEN** any one of the five policy values or the currency differs from the gate
- **THEN** the startup interlock refuses the Guardian as a single-source mismatch

#### Scenario: Risk budget derivation
- **WHEN** the Guardian policy is created from a valid gate
- **THEN** its per-trade risk budget equals the configured daily-loss amount and grants no larger loss budget

### Requirement: Production Guardian uses the engine journal
All decisions and reservations issued by the production Guardian SHALL be
written through the same `journal.Journal` instance owned and closed by the
engine context. Production assembly SHALL NOT open a second Guardian-only
journal.

#### Scenario: Guardian issuance storage
- **WHEN** the production Guardian issues an allowed decision
- **THEN** the decision and reservations are visible through the engine context's journal and share its account scope

#### Scenario: Context closes shared ownership
- **WHEN** the command-assembled context is closed after production Guardian issuance
- **THEN** the one engine-owned journal handle is closed and no Guardian-only journal remains open
