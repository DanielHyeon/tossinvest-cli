# Review — a066-add-multi-horizon-risk-buckets

- Date: 2026-08-03
- Stage: Wave 1A pure risk core hardening GREEN; journal/runtime integration pending
- Voices: Manager scope/safety review, independent adversarial risk review, final semantic re-review

## Findings and disposition

- **Accepted:** bucket dimensions are horizon/market/strategy/sector/symbol and canonical quantity fields
  are `q_candidate` and `q_final`.
- **Accepted:** HELD is a monetary reservation at worst executable price plus worst fees and fresh official
  FX haircut/ceil. `q_final` is fixed before final RiskIntent and GuardianDecision are issued atomically.
- **Accepted:** owner release requires the prior generation's mutations, protection, sell claims,
  reduce-only attempts and unresolved fills to be authoritatively clean.
- **Accepted after round 2:** each deduplicated fill transfers proportional HELD and records
  `filled=max(transferred conservative amount, actual monetary exposure)` in every applicable bucket.
  Overage or unknown actual price/fee/FX preserves the fill/Position and latches only new exposure as
  `RISK_OVERAGE` or `UNKNOWN_ACTUAL_RISK`.

## Verification

- Strict OpenSpec validation: PASS.
- Proposal-level semantic re-review: PASS. The first implementation review returned request-changes;
  all six findings below now have focused regression coverage and await independent re-review.
- Property/crash tasks cover partial fills, replacements, predecessor late fills, retries and atomic replay.

## Wave 1A implementation evidence

- Added the new leaf package `internal/riskbucket` only; existing journal, Guardian, Gateway and
  strategy-engine runtime functions were not edited by this wave.
- `ReservationMinor` and `MaximumQuantity` use exact decimal arithmetic, official frozen fresh
  price/FX evidence, FX haircut, minimum/nonlinear fee policy and account-minor-unit ceil with a
  bounded monotone integer search.
- `CalculateAdmission` normalizes horizon/market/strategy/sector/symbol order and records
  `q_candidate`, existing Guardian cap, `q_final`, complete policy preimage, bucket snapshot versions,
  binding caps and per-bucket reservations.
- `ApplyFill` is a crash-pure deep-copy transition with cumulative/fill identity watermarks,
  proportional HELD transfer, persisted actual price/fee/FX evidence, `max(transfer, actual)`,
  monotonic late evidence completion and all-applicable-bucket/owner entry latches. Unknown actual
  and overage preserve the authoritative fill watermark.
- Owner acquisition enforces one account/market/symbol owner even across competing prospective
  tokens. Actual generation binding is set-once. Release is idempotent and requires the complete
  CLOSED, zero, reconciliation, protection, sell/reduce-only and unresolved-fill clean predicate.
- Independent implementation-review hardening is implemented in the same leaf package:
  - every stored-minor addition and overage recomputation is 256-bit bounded, parse/overflow errors
    fail closed, and the returned/input state and fill watermark remain unchanged;
  - immutable typed policy provenance binds the exact `BucketKey`, while snapshot provenance binds
    key, version, limit, FILLED and HELD and is accepted only from the matching official frozen
    authority source within its freshness window;
  - owner release evidence and its fresh immutable attestation bind the exact owner key,
    prospective generation, lane, campaign and actual generation;
  - actual-fill FX evidence binds the order quote/base currency pair; mismatch preserves the fill,
    transfers conservative HELD, and latches `UNKNOWN_ACTUAL_RISK`;
  - duplicate account/market/symbol owners always return `RECONSTRUCTION_MISMATCH`, independent of
    Go map iteration order; entry blocking consults both owner aggregate and bucket-local latches.
- Function Logic Map: `not-applicable` for Wave 1A because every implementation function is new and
  no existing function body was changed. The executable table/property/fuzz tests are the Branch
  Test Map for this new leaf package.
- RED evidence: `go test ./internal/riskbucket` initially failed to compile because the new
  reservation, admission, fill and owner contracts had no implementation.
- Review-fix RED evidence: focused tests initially failed to compile on the missing provenance,
  release-attestation and currency-pair contracts.
- Review-fix GREEN evidence: `go test -race ./internal/riskbucket`, `go vet
  ./internal/riskbucket`, 25 repeated property/reconstruction/overflow runs, two focused 3-second
  fuzz runs and `git diff --check` passed. The fuzz runs executed 475,508 reservation cases and
  222,136 fill-retry cases. Statement coverage at this checkpoint is 77.6%.
- `make sdd-sync` completed the CodeGraph phase (`27 changed files`, then 1,368 indexed files,
  23,745 nodes and 77,709 edges). The advisory CodeGraphContext update stalled after database load
  and was terminated instead of delaying the focused Wave 1A work.

## Deferred boundary

Wave 1B still owns additive journal schema/replay, atomic decision+all-bucket reservation commits,
Guardian/Gateway wiring, KR/US concurrent runtime integration, cancel/expiry/restart reconstruction,
entry-only loss-lock and risk-reducing bypass tests. No live order, operating toggle or automation
activation was performed. Independent implementation review and the full a066 gate remain pending.

## Verdict

Proposal freeze remains approved. Wave 1A pure core is GREEN, but the change is not production-ready
until Wave 1B integration and the full gate complete. Missing official FX evidence yields zero
exposure-raising quantity and must never delay fill, reconciliation, protection or reduce-only exit.
