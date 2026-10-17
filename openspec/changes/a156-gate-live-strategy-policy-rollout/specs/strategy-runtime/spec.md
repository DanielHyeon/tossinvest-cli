## ADDED Requirements

### Requirement: Live rollout is separately authorized and evidence bound
Live enablement SHALL require completed predecessor evidence, independently reviewed versioned manifests, bounded risk and market scope, rollback procedures and explicit human activation approval; code or design completion SHALL NOT activate trading.

#### Scenario: Approval absent
- **WHEN** all offline checks pass but live approval is absent
- **THEN** new policy remains inactive and the change cannot claim live completion

### Requirement: Later legs and restarts preserve campaign protection
Each later-leg admission SHALL recompute bounded quantity from actual fills, outstanding reservations and nonretreating protection; uncertain broker state SHALL block new exposure until authoritative reconciliation.

#### Scenario: Ambiguous cancellation
- **WHEN** cancel returns 404 or an old order is absent from one snapshot
- **THEN** no terminal cancellation is assumed and no duplicate replacement entry is sent

### Requirement: Intraday flattening is scoped and fill verified
Session-end closeout SHALL target only authorized intraday campaign inventory using exchange deadlines and SHALL require authoritative residual and fill confirmation before declaring flat.

#### Scenario: Weekly holding coexists
- **WHEN** intraday closeout runs while weekly or adopted inventory remains
- **THEN** unrelated inventory is preserved and pending/partial exits are reported as unconfirmed

### Requirement: Rollback preserves exits and historical state
Rollback SHALL stop new policy entries while keeping current protection operational and historical policy snapshots intact.

#### Scenario: Rollout aborted
- **WHEN** the new policy rollout is stopped during an open position
- **THEN** existing protective exits continue and no historic state is overwritten
