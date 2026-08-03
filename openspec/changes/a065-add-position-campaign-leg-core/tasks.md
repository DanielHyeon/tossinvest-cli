## 1. Pre-Edit Evidence and Logic Maps

- [ ] 1.1 Run `make sdd-sync`, record CodeGraph definitions/callers/callees/impact for journal migrations, tx-scoped fill apply, Position projection, exit state and reconciliation, and pin the current change base commit.
- [ ] 1.2 Complete Go AST artifacts, Function Logic Maps, Branch Test Maps and risk-pattern reports for every existing fill/journal/Position/exit function before editing it, including crash, duplicate and EXIT FIRST branches.
- [ ] 1.3 Freeze design D4's complete Campaign/Leg transition tables, prospective-generation CAS, per-order watermark/replacement scheme, command keys, event ordering, stop monotonicity and Position lineage as executable fixtures.

## 2. RED Contract Tests

- [ ] 2.1 Add failing domain tests for strategy-neutral campaign/leg identities, ordered sequence, new generation after CLOSED and rejection of embedded 8:4:2, 2:4:8 or seven-leg constants.
- [ ] 2.2 Add failing command tests for prospective-generation CAS races, retry idempotence, concurrent expected-version conflict and all-or-nothing first-fill binding crash boundaries.
- [ ] 2.3 Add failing transition tests for every Campaign/Leg state-table row, EXIT FIRST races, RECONCILE observation handling, CLOSED terminality and invalid-transition isolation.
- [ ] 2.4 Add failing stop tests for saved-stop monotonicity, invalid/missing candidates and source/policy/observed-at provenance.
- [ ] 2.5 Add failing per-order watermark tests for duplicate/lower cumulative observations, amend/replacement carry baselines, and replaced/cancelled predecessor late positive deltas; prove retry/restart is delta-zero, commit crash is all-or-nothing, replacement remaining/leg aggregate is recalculated, and cap excess or ambiguous lineage preserves fill/Position while latching campaign RECONCILE and entry block.
- [ ] 2.6 Add failing offline reconstruction tests for prospective binding, restart parity, sequence gaps, duplicate keys, orphan order lineage, snapshot drift and the prohibition on automatic repair or broker calls.

## 3. Additive Journal Schema

- [ ] 3.1 Add an atomic additive migration for campaign/leg events, prospective-generation tokens, per-order watermarks, projections and nullable lineage with uniqueness, foreign-key and sequence constraints.
- [ ] 3.2 Update schema golden tests for migration atomicity, legacy campaign-unknown reads and ErrSchemaTooNew without synthesizing campaign IDs for existing Positions.
- [ ] 3.3 Implement journal transaction primitives for prospective position generation CAS, expected campaign version, deterministic command key, per-order watermark, event append and projection update with idempotent result retrieval.

## 4. Campaign Core

- [ ] 4.1 Implement strategy-neutral `PositionCampaign`, `CampaignLeg` and order-watermark types plus every row of the frozen Campaign/Leg state-transition tables.
- [ ] 4.2 Implement idempotent prospective campaign creation, plan, submit-link, cancel and broker-order cumulative fill commands without broker submission or lane-specific quantities/cadence.
- [ ] 4.3 Implement EXIT FIRST admission so EXITING/CLOSING/RECONCILE and unresolved risk-reducing intent reject new exposure while fill detection, reconciliation and emergency exits continue.
- [ ] 4.4 Implement long-only effective-stop composition as `max(saved, valid candidate)` with immutable candidate and selection provenance and no protection mutation.
- [ ] 4.5 Implement read-only event replay of prospective binding and per-order replacement/watermark lineage plus snapshot comparison that reports stable mismatch reasons and last valid event without changing journal or runtime state.

## 5. Position and Fill Integration

- [ ] 5.1 Extend explicit journal lineage from prospective generation through Campaign, Leg and order watermark to decision, intent, mutation attempt, Fill and Position generation while keeping Position quantity/average price as the sole authority.
- [ ] 5.2 Integrate first-fill prospective-token binding and per-order watermark/event application into the existing tx-scoped fill hook so fill snapshot, Position delta, exit state and campaign projection commit or rollback together.
- [ ] 5.3 Preserve replaced/cancelled predecessor late positive deltas by advancing its immutable watermark and Position exactly once in the fill transaction, recalculating successor remaining and leg aggregates, and latching campaign RECONCILE/new-entry block without truncating over-cap or ambiguous-lineage fills.
- [ ] 5.4 Add crash/restart and race integration tests covering concurrent campaign creation, immediate full fill, repeated/replacement partial fill, residual cancel, predecessor late terminal fill, cap excess, ambiguous replacement lineage, exit-versus-scale-in and every RECONCILE recovery row.
- [ ] 5.5 Expose only offline reconstruction/read models; leave all production entry callers, live dispatch and lane/automation activation disconnected.

## 6. Verify and Gate

- [ ] 6.1 Run focused unit, migration, race and journal integration suites and record RED-to-GREEN evidence for every transition-table and Branch Test Map row.
- [ ] 6.2 Run broker spies and configuration assertions proving campaign planning/replay emits zero live requests, does not flip lane/automation toggles and never delays stop, emergency exit, reconciliation or fill detection.
- [ ] 6.3 Refresh Function Logic Maps, Branch Test Maps and risk reports after edits, then run `openspec validate a065-add-position-campaign-leg-core --strict --no-interactive`, `make sdd-check`, `make test`, `make vet` and `make validate`.
- [ ] 6.4 Complete adversarial independent review for journal atomicity, EXIT FIRST and non-retreating stops, resolve findings and run `make gate CHANGE=a065-add-position-campaign-leg-core` with live entry callers still absent.
