## ADDED Requirements

### Requirement: Initial protection and sizing share a versioned campaign contract
New supported campaign admission SHALL bind initial stop, price basis, risk budget, final quantity, source and policy identity through execution terms, reservation, journal and exit initialization; pre-submit mismatches SHALL block new entry while existing protective exits continue. Post-fill initialization mismatches SHALL preserve the admitted protection snapshot, block further exposure and enter the existing protection-recovery/reconciliation path with an explicit alarm; they SHALL NOT be reported as a retroactively refused fill.

#### Scenario: Stop mismatch
- **WHEN** before order submission an exit initialization contract stop differs from the admitted initial stop without an authorized tightening transition
- **THEN** new entry is refused with an explicit reason and existing protection is not disabled


#### Scenario: Mismatch after fill
- **WHEN** a fill is confirmed before exit initialization detects a mismatch
- **THEN** the fill remains recorded, additional campaign entry is blocked, the admitted stop is retained for existing protection recovery, and unresolved protection is reported explicitly

### Requirement: Quantity respects all existing caps and risk bases
Sizing SHALL retain all Guardian and bucket caps, reject nonpositive per-share risk and treat a zero cap as zero admissible quantity. Planned and filled risk SHALL be reported separately.

#### Scenario: Zero notional cap
- **WHEN** risk budget allows shares but a mandatory notional cap permits zero
- **THEN** final admitted quantity is zero and no order is dispatched

### Requirement: Existing campaigns preserve their protection snapshots
Policy changes SHALL apply only to new campaigns and SHALL NOT widen existing stops, rewrite historical rows or activate later legs.

#### Scenario: Policy version changes
- **WHEN** a new initial-stop policy is selected while an old campaign exists
- **THEN** old campaign keeps its bound version and stop; new version requires fresh admission
