## ADDED Requirements

### Requirement: External order provenance remains stable across cycles

An order classified as external because no local intent, confirmed mutation attempt, or lineage owns it SHALL remain external across polling cycles. Persisting a broker observation MUST NOT by itself turn that order into local open exposure, and its later absence MUST NOT create a missing-local-order mismatch or contribute to permanent promotion.

#### Scenario: Stored external observation is absent later
- **WHEN** reconciliation runs after a previously stored external open-order observation disappears from the broker open list
- **THEN** the corrected local state contains no such open order, the diff contains no missing order for it, and the tracker failure counter does not increase because of it

### Requirement: Reconcile release is durable before it is visible

The tracker SHALL preserve an existing block and its entry-gate projection until the corresponding journal release commits. If persistence fails, the current reconciliation cycle MUST stop before automatic adoption. Permanent release SHALL require explicit operator identity, explanatory evidence, engine exclusion, and a fresh stable authoritative comparison with no blocking diff.

#### Scenario: Durable release fails
- **WHEN** a matching post-adjustment comparison proposes a release but the journal write fails
- **THEN** the tracker and gate remain blocked and the cycle performs no adoption or price read

#### Scenario: Operator resolves a proven-clean account
- **WHEN** the engine is stopped, three official snapshots taken at least two seconds apart agree, corrected local state has no blocking diff, and the operator explicitly confirms a release with identity and note
- **THEN** supported active states are released atomically with `OPERATOR` evidence before the tracker/gate are cleared

#### Scenario: Operator release sees a blocking diff
- **WHEN** the fresh corrected comparison still contains a quantity mismatch or locally owned missing order
- **THEN** the release is refused and every active journal state remains unchanged

#### Scenario: Prior-session lineage could match the fresh broker snapshot
- **WHEN** a prior account session has a replacement child whose identifier appears in the fresh selected-account snapshot
- **THEN** the corrected local state keeps the current-session parent distinct and the recovery command refuses release if that current order is missing

### Requirement: Startup reservation recovery preserves canonical ownership

Before the engine can take a decision or prune spent-nonce evidence, startup SHALL run reservation recovery. The sweep SHALL release a held reservation from a terminal fill snapshot only when the reservation's decision-bound confirmed attempt and intent exactly match the snapshot account, market, market-local trading day, symbol, and side, and that canonical scope has exactly one intent owner. A cross-scope or ambiguous snapshot MUST leave risk headroom held; ambiguity MUST record an account-wide identifier conflict. Spent nonce evidence referenced by a held reservation MUST remain retained regardless of the ordinary age cutoff.

#### Scenario: Reused order identifier has a terminal snapshot in another scope
- **WHEN** a held reservation's order identifier has a terminal snapshot from another account, market, trading day, symbol, or side
- **THEN** startup recovery releases nothing and the reservation remains held

#### Scenario: Old spent nonce still protects a held reservation
- **WHEN** a spent nonce is older than the retention cutoff but its decision still owns a held reservation
- **THEN** retention preserves the nonce and a later startup cannot release the hold as expired-unconsumed
