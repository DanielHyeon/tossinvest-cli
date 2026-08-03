## 1. Pre-Edit Evidence and Logic Maps

- [ ] 1.1 Run `make sdd-sync`, record CodeGraph definitions/callers/callees/impact for Guardian evaluation/issuance, reservation transactions, Gateway revalidation, fill apply and mode/loss locks, and pin the current change base commit.
- [ ] 1.2 Complete Go AST artifacts, Function Logic Maps, Branch Test Maps and risk-pattern reports for every existing Guardian, reservation, Gateway and fill function before editing it, including all fail-closed and risk-reducing bypass branches.
- [x] 1.3 Freeze horizon/market/strategy/sector/symbol key ordering, canonical q_candidate/q_final fields, monetary reserve formula, prospective/actual owner lifecycle, protection/sell-clean release, held-to-filled transitions and refusal codes as executable tables.

## 2. RED Contract Tests

- [x] 2.1 Add failing pure calculator/property tests proving `0 <= q_final <= q_candidate` and `q_final <= existing Guardian cap` across horizon, market, strategy, sector and symbol monetary boundaries.
- [x] 2.2 Add failing monetary reservation tests for worst executable price, nonlinear/minimum fees, FX rate/haircut >= 1, minor-unit ceil, same-currency identity, stale/missing provenance and decimal overflow.
- [x] 2.3 Add failing reservation tests for concurrent prospective-generation ownership, atomic all-bucket acquisition, q_final-before-decision ordering, partial rollback prevention and same-owner scale-in.
- [x] 2.4 Add failing held/filled accounting tests for partial, replacement and predecessor-late fill; require per-fill proportional HELD transfer versus actual price/fee/FX exposure, `filled=max(transfer,actual)`, all-bucket overage, duplicate/retry idempotence, crash atomicity, cancel/expiry, restart replay, orphan reservation and snapshot drift.
- [x] 2.5 Add failing safety tests proving cap overage or unknown actual price/fee/FX never drops/truncates/rolls back a fill or Position, latches every applicable bucket/owner with `RISK_OVERAGE` or `UNKNOWN_ACTUAL_RISK`, and blocks new exposure only.
- [x] 2.6 Add failing owner-release tests proving CLOSED alone is insufficient until reconciliation, prior protection order/saga, sell/reduce-only claim/mutation and unresolved fill evidence are all clean.
- [ ] 2.7 Add failing safety tests proving horizon/market loss locks and all bucket failures block exposure-raising only and cannot delay stop, emergency exit, reconciliation or fill detection.

## 3. Additive Journal Schema

- [x] 3.1 Add an atomic additive migration for monetary bucket policies/snapshots, strategy keys, prospective/actual lane owners, reservations and events with explicit price/fee/FX provenance. Preserve released v22 exactly; place incompatible scoped order/fill shapes in a backed-up, transactional v23 transition that retains v22 tables under immutable legacy names without promotion.
- [x] 3.2 Update schema golden tests for migration atomicity, legacy unknown state and ErrSchemaTooNew without defaulting missing exposure, FX, sector or ownership.
- [x] 3.3 Implement journal replay/read primitives that reconstruct owner, HELD/filled usage and digest, returning stable mismatches without automatic owner replacement or reservation deletion.

## 4. Risk Bucket Core

- [x] 4.1 Implement typed horizon/market/strategy/sector/symbol bucket keys, normalized acquisition order and versioned monetary policy snapshots with explicit unknown states.
- [x] 4.2 Implement the exact conservative monetary reserve function and maximum-integer cap search over worst price, fees, FX haircut and minor-unit ceil.
- [x] 4.3 Implement q_final as the minimum of q_candidate, existing Guardian and every monetary bucket cap with typed zero-quantity refusals and complete preimage.
- [x] 4.4 Implement tx-scoped monetary usage accounting for every applicable bucket: proportional HELD transfer, persisted actual price/fee/FX exposure, `filled=max(transfer,actual)`, monotonic actual-evidence completion, and durable all-bucket overage/unknown latches without rejecting authoritative fills. (Pure transition plus owner-wide multi-decision authoritative journal accounting are GREEN; production actual-evidence authority remains part of later integration.)
- [ ] 4.5 Implement prospective-to-actual unique ownership binding and idempotent release only after CLOSED, broker-zero reconciliation and prior protection/sell claim cleanliness. (Journal lifecycle hardening is independently CLEAN, but official broker-zero capability deliberately has no production mint/caller until an immutable official holdings adapter lands.)

## 5. Guardian and Gateway Integration

- [ ] 5.1 Split Guardian evaluation into mutation-free precheck/cap calculation and a final issuance step that occurs only after q_final is known.
- [ ] 5.2 Extend the final transaction to commit the q_final GuardianDecision, existing reservation, all monetary bucket reservations and lane owner atomically, rolling everything back on any conflict.
- [ ] 5.3 Extend Guardian provenance and stable reason codes with q_candidate, q_final, binding monetary cap, price/fee/FX digest, bucket versions and owner identity.
- [ ] 5.4 Extend Gateway pre-submit revalidation to require decision quantity equals q_final plus every HELD monetary reservation and matching prospective/actual owner before any broker request.
- [ ] 5.5 Implement short/medium and market-specific entry loss locks through an entry-only port, with durable conservative activation and human-approved audited relaxation.
- [ ] 5.6 Add KR/US concurrent integration tests proving independent market buckets, shared strategy caps, one symbol/one owner, monetary scale-in aggregation and failure isolation without requiring an operating toggle.
- [ ] 5.7 Add partial/replacement/predecessor-late-fill crash and retry integration tests proving Position, watermark, proportional transfer, filled max and every bucket overage/latch commit exactly once while risk-reducing paths remain available.

## 6. Verify and Gate

- [ ] 6.1 Run monetary calculator property tests, migration/journal integration tests and races for prospective owner, q_final decision ordering, all-bucket acquisition and protected release; record RED-to-GREEN evidence for every Branch Test Map row.
- [ ] 6.2 Run broker spies and timing assertions proving all bucket/loss-lock failures cause zero exposure-raising live requests while risk-reducing, stop, emergency exit, reconciliation and fill detection remain callable without evidence/FX waits.
- [ ] 6.3 Verify the feature is dormant by default, does not loosen any existing limit, does not flip lane/automation/live toggles and treats unresolved US FX as q_final 0.
- [ ] 6.4 Refresh Function Logic Maps, Branch Test Maps and risk reports after edits, then run `openspec validate a066-add-multi-horizon-risk-buckets --strict --no-interactive`, `make sdd-check`, `make test`, `make vet` and `make validate`.
- [ ] 6.5 Complete adversarial independent review for reservation atomicity, quantity monotonicity, single-owner races and exit bypass, resolve findings and run `make gate CHANGE=a066-add-multi-horizon-risk-buckets` without activating live trading.
