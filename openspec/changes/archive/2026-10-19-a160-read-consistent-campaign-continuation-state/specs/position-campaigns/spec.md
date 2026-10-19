## ADDED Requirements

### Requirement: Continuation inspection reads a single non-authorizing campaign snapshot
The query-only journal handle SHALL read the requested account, market, symbol, campaign and expected actual position generation in one database read transaction. It SHALL return the current position and projection version, campaign state and protection facts, ordered legs and order watermarks without granting execution authority.

#### Scenario: Concurrent fill cannot mix snapshot generations
- **WHEN** a separate writer commits a fill and associated projection updates after the inspection establishes its database snapshot
- **THEN** all inspected records belong to the established snapshot and a later inspection can observe the committed state without mixing the two

#### Scenario: Scope or current generation cannot be proven
- **WHEN** request scope differs, the latest position generation differs, authoritative projection version is missing, the nonempty position entry decision differs from the campaign decision, or campaign-position linkage is unavailable
- **THEN** inspection returns a typed error or unavailable result without a usable partial snapshot, zero-risk fallback or new rows

#### Scenario: Protection blocks entry but remains inspectable
- **WHEN** a matching current open position has a reconciled or entry-blocked campaign
- **THEN** the snapshot preserves the stored block and protection facts without clearing them or refusing solely because entry is blocked

### Requirement: Missing continuation plan evidence remains explicitly unavailable
Every successful continuation inspection SHALL explicitly mark full immutable campaign plan and budget preimage unavailable under the current schema. It SHALL NOT infer that preimage from a digest, leg quantity or summed monetary buckets, and SHALL NOT return a ready-for-entry flag, new quantity decision, reservation or lease.

#### Scenario: Existing rows do not prove a complete immutable plan
- **WHEN** a valid current campaign has readable position, legs and watermark facts
- **THEN** the result still carries the stable full-plan-unavailable reason and creates no authority or reconstructed zero risk

### Requirement: Continuation inspection preserves journal contents
Continuation inspection SHALL use the existing read-only handle without schema changes or repairs. Missing required legs, orphan watermarks and read failures SHALL fail without a usable partial result; a valid unfilled planned leg MAY have no watermark. Reading SHALL leave logical journal rows unchanged.

#### Scenario: Reopen preserves inspection facts
- **WHEN** the same unchanged synthetic journal is closed and reopened read-only
- **THEN** inspection returns the same stored facts and unavailable-plan reason with no change to decisions, reservations, leases, protection or campaign rows
