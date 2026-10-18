## ADDED Requirements

### Requirement: Initial protection and sizing share a versioned campaign contract
New supported campaign admission SHALL bind initial stop, price basis, risk budget, final quantity, source and policy identity through execution terms, reservation, journal and exit initialization; pre-submit mismatches SHALL block new entry while existing protective exits continue. Post-fill initialization mismatches SHALL preserve the admitted protection snapshot, block further exposure and enter the existing protection-recovery/reconciliation path with an explicit alarm; they SHALL NOT be reported as a retroactively refused fill.

#### Scenario: Stop mismatch
- **WHEN** before order submission an exit initialization contract stop differs from the admitted initial stop without an authorized tightening transition
- **THEN** new entry is refused with an explicit reason and existing protection is not disabled


#### Scenario: Mismatch after fill
- **WHEN** a fill is confirmed before exit initialization detects a mismatch
- **THEN** the fill remains recorded, additional campaign entry is blocked, the admitted stop is retained for existing protection recovery, and unresolved protection is reported explicitly

#### Scenario: Missing new-version binding
- **WHEN** a decision marked as requiring the new protection contract has a missing or unreadable binding
- **THEN** it does not fall back to legacy admission, new authorization is refused, and any already-existing protection is preserved

#### Scenario: Already authorized transport races with a latch
- **WHEN** a command has passed the final consistent authority read before a campaign latch is persisted
- **THEN** it remains possibly in-flight with reservations and reconciliation preserved rather than being reported as unsent or cancelled; subsequent authority reads after the latch refuse new authorization

#### Scenario: Latch persistence fails
- **WHEN** protection reconciliation cannot persist its campaign latch
- **THEN** the failure is reported and retried without suppressing valid admitted protection, and final transport authorization independently refuses unreadable or conflicting durable protection evidence

### Requirement: Quantity respects all existing caps and risk bases
Sizing SHALL retain all Guardian and bucket caps, reject nonpositive per-share risk and treat a zero cap as zero admissible quantity. Planned and filled risk SHALL be reported separately.

#### Scenario: Zero notional cap
- **WHEN** risk budget allows shares but a mandatory notional cap permits zero
- **THEN** final admitted quantity is zero and no order is dispatched

### Requirement: Admitted protection floor recovery preserves concurrent protection
For a verified same-generation admission, recovery SHALL preserve a tighter existing baseline and historical initial risk provenance. Raising a weaker baseline SHALL first persist a compatible evaluated snapshot without cancelling or submitting orders, using an explicit floor-reconciliation suppression reason. The next real observation SHALL use the durable floor through the existing order path. Before selecting a recovery snapshot, the same journal transaction SHALL compare the observed effective snapshot ID, baseline, high-water, level/rung, current quantity, taken ratio and pending proposal fields, retaining existing lifecycle and generation checks. A stale or incompatible candidate SHALL leave existing protection unchanged rather than quarantine it.

#### Scenario: Initial snapshot is absent and another writer advances state
- **WHEN** a floor candidate was evaluated before a concurrent writer changes watermark, quantity or pending proposal while the effective snapshot ID remains empty
- **THEN** its state comparison fails before recovery selection, no order is cancelled or submitted, and the newer protection remains intact for the next observation

#### Scenario: Breached floor survives restart
- **WHEN** a floor-reconciliation snapshot is committed while price breaches the admitted stop and the process restarts before another observation
- **THEN** existing protection and pending orders remain preserved, the next real observation uses the durable floor without repeated floor suppression, and any resulting order retains authoritative sellable limits

#### Scenario: Foreign generation cannot recover from an old admission
- **WHEN** the position generation differs from the verified admission projection
- **THEN** that admission is not returned as a usable seed or floor for the foreign position, while already-existing protection remains available

### Requirement: Existing campaigns preserve their protection snapshots
Policy changes SHALL apply only to new campaigns and SHALL NOT widen existing stops, rewrite historical rows or activate later legs.

#### Scenario: Policy version changes
- **WHEN** a new initial-stop policy is selected while an old campaign exists
- **THEN** old campaign keeps its bound version and stop; new version requires fresh admission
