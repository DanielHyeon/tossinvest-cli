# Review — a066-add-multi-horizon-risk-buckets

- Date: 2026-08-04
- Stage: Wave 1C single-decision journal accounting GREEN; runtime and aggregate integration pending
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

## Wave 1B journal checkpoint

- Schema v22 is additive. Immutable authority policy/snapshot rows retain explicit worst-price,
  fee and FX source/version/digest/freshness fields; v21 history remains bucket-state-unknown.
- `CommitRiskBucketAdmission` reruns the pure calculator inside the journal boundary, checks exact
  snapshot key/version/digest bindings and the referenced HELD legacy reservation, then commits the
  final quantity, owner, five reservations, event and replay digest in one SQLite transaction.
- RED was the expected compile failure for the absent v22 symbols and admission API. GREEN covers
  migration rollback, no-backfill, exact retry, partial-write rollback, two-process owner races,
  orphan references, snapshot-version drift and stable state-digest mismatch without repair.
- No existing function body changed. The only existing-code edit is the declarative schema version
  and appended migration entry, so a pre-edit Function Logic Map was not applicable.
- This checkpoint does not implement authoritative fill/release transactions or actual
  Guardian/Gateway integration. Tasks 2.4, 2.5, 5.x and the full gate remain pending.

### Independent source-review hardening

- Immutable policy/snapshot inserts now re-read and compare a full-record digest. A reused primary
  key, snapshot ID or unique digest with different amount/provenance fails the entire transaction;
  `INSERT OR IGNORE` can no longer bless mismatched authority evidence.
- Active-owner reuse now requires exact prospective generation, lane and campaign identity, matching
  the pure owner contract. Same lane/campaign with a different prospective token is a conflict.
- Same-owner scale-in receives a transaction-scoped monotonic owner sequence. Commit and replay share
  one DB-derived canonical preimage over ordered decisions and reservations, snapshot IDs, owner and
  scope latches, HELD/FILLED/overage/state fields; aggregate quantity and monetary usage are exact and
  bounded. RFC3339 text ordering is not used for authority.
- Added regressions for immutable collisions, prospective-token conflict, two-step scale-in,
  reservation deletion, snapshot rebinding and field tamper. Every mismatch remains fail-closed and
  leaves persisted evidence untouched.
- Admission receipts are pinned to exactly five unique reservation IDs. Scale-in is allowed only for
  the exact same five bucket keys and policy versions; a strategy/key/version change requires a new
  owner lifecycle and is rejected before any decision row is inserted.
- Owner identity is now cross-bound to the authoritative market and symbol buckets. A KR owner with
  a US market bucket, or any owner/symbol-bucket mismatch, is rejected before the transaction begins.
- `BucketSnapshot.BoundEvidence` returns value copies of the already sealed private provenance only;
  it cannot construct or mutate an authority seal. Journal validation uses it to require exact
  policy/snapshot source, version, digest, observed/fresh times and bound amounts before writing.
- The idempotence preimage includes the canonical ordered full consumed bucket bindings, not merely
  their computed availability/caps. A retry that preserves `available` and `q_final` while changing
  limit/FILLED/HELD or sealed evidence is a divergent replay, never an idempotent success.
- Focused tests and their race run, journal vet, strict OpenSpec validation and diff whitespace check
  pass. The single full journal run produced no output before its explicit 240-second timeout, so it
  is recorded as incomplete rather than reported as passing.

## Verdict

Proposal freeze remains approved. Wave 1A pure core is GREEN, but the change is not production-ready
until Wave 1B integration and the full gate complete. Missing official FX evidence yields zero
exposure-raising quantity and must never delay fill, reconciliation, protection or reduce-only exit.

## Wave 1C authoritative fill checkpoint

- The only existing fill-path body edit is a tx-scoped a066 sidecar call in `Journal.RecordFill`.
  It executes before the existing Position/campaign/exit hooks and therefore shares their commit or
  rollback boundary.
- Focused RED-to-GREEN coverage proves partial/replacement/predecessor-late fills, duplicate actual
  completion, cancel/expiry release, outer-hook rollback, restart, orphan mapping, state drift and
  risk-reducing bypass. Unknown or over-limit actual exposure preserves the authoritative fill and
  latches all five buckets plus the owner.
- Review findings resolved in this wave: monetary aggregation uses bounded 256-bit arithmetic;
  actual/release commands require exact owner, decision, account, market and order identity; active
  registered orders prevent later scale-in admission; and a replaced predecessor cannot release the
  HELD reservation already handed to its successor.
- Follow-up adversarial findings resolved: a predecessor must be the exact ACTIVE decision order and
  its transition must affect exactly one row; terminal or already-REPLACED parents cannot seed another
  child. Release replay requires the original reason. Order/fill digests use canonical required-order
  slices with marshal errors propagated, and policy/currency authority is derived from all five sealed
  persisted policies rather than caller-controlled strings.
- Ambiguous/corrupt sidecar state has an explicit non-drop path: every applicable reservation and owner
  is conservatively latched and `FILL_UNACCOUNTED` is appended while the authoritative fill and Position
  commit. Database transport errors remain outer-transaction failures.
- Post-review CRITICAL — Wave 1C had changed the already released v22 table shapes in place. Resolved by
  restoring `schemaV22` exactly to commit `4aee6853`, incrementing `SchemaVersion` to 23 and preserving
  legacy order/fill/allocation rows in immutable `_v22` tables. They are not auto-promoted because v22
  lacks the scoped identity and evidence required by the new authority model. Migration failure rolls
  back every rename and `user_version`; an older v22 build refuses the v23 journal.
- Post-review authority boundary — caller-created `Official/Frozen` flags and CANCEL/EXPIRY enums are not
  production capabilities. Actual completion and release methods are package-private, have a static
  zero-production-caller guard and remain pending official sealed evidence plus journal-derived cancel,
  expiry, broker-zero and clean lifecycle validation.
- Final independent review found the registered order quantity was caller-supplied even after broker-order
  authority had been confirmed. Registration now derives the exact confirmed intent quantity in the same
  transaction and refuses a missing, ambiguous, non-integral or divergent quantity before writing an order.
- Remaining limitation: the journal adapter deliberately rejects owners with multiple final
  decisions rather than guessing an aggregate binding. That aggregate model, actual owner binding,
  clean owner release and Guardian/Gateway runtime wiring remain required before production use.
