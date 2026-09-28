# Review — a066-add-multi-horizon-risk-buckets

- Date: 2026-08-04
- Stage: Wave 1E v24 owner-lifecycle BLOCK fixes GREEN; independent re-review CLEAN
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
  all six findings below now have focused regression coverage and independent re-review is CLEAN.
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
- Wave 1C limitation: the journal adapter deliberately rejected owners with multiple final
  decisions rather than guessing an aggregate binding. Wave 1D resolves that aggregate model;
  actual owner binding, clean owner release and Guardian/Gateway runtime wiring remain required
  before production use.

## Wave 1D owner-wide aggregate fill checkpoint

- RED pinned the former active-order scale-in refusal: a second exact decision failed with
  `scale-in while risk order accounting is active`. GREEN removes only that single-decision guard;
  owner/key/policy drift still fails before a decision write.
- Each confirmed order is bound to its exact decision and immutable internal `order_key`. Owner-wide
  reconstruction includes every decision/order/fill/allocation, refuses a broker-ID collision and
  applies aggregate monetary deltas only to the target order's five reservation IDs.
- The pure fill transition accepts `OrderKey` and a target-decision HELD view. This prevents one
  scale-in decision's late fill from consuming another decision's HELD; any deficiency becomes a
  durable conservative overage latch without changing the authoritative fill watermark.
- Two decision-specific partial fills produce exact aggregate HELD/FILLED values and zero
  cross-decision allocations. Restart reconstruction is stable, and late actual completion clears
  UNKNOWN only after all owner fills have authoritative evidence.
- Corrupt sidecar identity still commits both authoritative fill and Position and latches all owner
  reservations with `FILL_UNACCOUNTED`. A confirmed ownership conflict now also latches every
  registered owner decision in matching scope while preserving the fill ledger.
- Actual-evidence completion and release APIs remain package-private. No schema migration/version,
  runtime toggle, Gateway, broker or live-order behavior changed in this checkpoint.

## Wave 1E authoritative owner lifecycle checkpoint — v24 hardening

- `runApplyHooks` now derives and binds the a066 owner only after successful PositionCampaign apply,
  inside the existing authoritative fill transaction. KR and US use one contract with exact market
  identity; no sequential "KR first, US later" dependency was introduced.
- Generation, CLOSED/zero, entry decision, campaign/claim, HELD/order, protection saga/attempt,
  BUY/SELL mutation, fill actual/latch and reconciliation facts are read from journal rows. Callers
  cannot authorize lifecycle changes with booleans, enums, generations or attestations.
- Broker-zero authority is no longer freeform reconcile evidence. Additive v24 preserves released v23
  and stores one structured official observation keyed by exact account/market/symbol/actual generation,
  with fixed official source, canonical zero, broker-as-of, capability/build/source versions and payload
  digest. The reconcile release stores that observation ID and digest. Operator-only evidence is invalid.
- The prior scalar recorder was itself an authority fabrication seam. It is removed: the journal recorder
  accepts only an opaque sealed capability whose exact scope/time/provenance/payload are seal-bound. There is
  no production constructor or call site in this change; official zero therefore remains structurally
  unreachable until an immutable official holdings adapter owns the mint path.
- `ADJUSTMENT_APPLIED` binds the exact append-only zero adjustment digest and additionally requires a
  later fresh official zero recheck; adjustment alone never authorizes release.
- Dirty or stale release returns a typed blocking field and performs zero writes. Clean release writes
  one append-only event plus immutable receipt binding owner/generation, campaign/Position versions,
  observation, predecessor sequence/digest and release time. Retry validates and recomputes every seal;
  there is no early AlreadyReleased return for a missing/divergent event, receipt or current state.
- Semantic bind gaps latch future entry without returning an error to the fill path. Replay drift is
  not silently resealed. A full late `RecordFill` for the released predecessor writes ORPHAN_FILL plus
  market-scoped symbol reconcile evidence with or without a reopened owner. The admission gate precedes
  owner lookup/INSERT, so first fresh admission is refused; exact market scope prevents US evidence from
  contaminating the same account/symbol in KR.
- Function Logic Maps are current for `Journal.runApplyHooks`, `CommitRiskBucketAdmission` and the
  active-only `loadRiskBucketState` wrapper. New lifecycle/receipt helpers are leaf implementations.
- Focused owner/migration tests, focused race, journal vet and full journal unit suite pass. Task 4.5
  remains unchecked until independent re-review. No runtime toggle, order dispatch, stop or emergency-exit
  path was added or delayed.

### Wave 1E follow-up — scoped reconciliation isolation

- Reviewer BLOCK accepted: retaining `idx_reconcile_active(account_ref,symbol)` made `scope_market`
  decorative and prevented simultaneous KR/US guards. v24 now drops it, preserves the account-wide index,
  and creates exact global-NULL and `(account_ref,symbol,scope_market)` active indexes.
- Insert/update overlap triggers preserve legacy NULL as global authority in both directions. API entry also
  searches global-or-exact while release selects exact only, so a KR release cannot release US or global.
- Late-fill insertion now blocks only global NULL or the same exact market. A reverse-order regression seeds
  KR first, creates US from the released-owner late fill, rejects same-market duplication, and proves KR/US
  admission and release isolation.
- `ReconcileState`, enter requests and single/atomic release requests carry normalized KR/US scope. Active
  reads expose it, IDs bind it, batch dedup includes it, and invalid/account-wide market combinations fail
  closed before a transaction.
- Verification: focused tests PASS; full `go test ./internal/journal -count=1` PASS (163.752s); focused race
  PASS (20.447s); `go vet ./internal/journal`, strict OpenSpec validation and `git diff --check` PASS.
- Final independent re-review: CLEAN with zero Critical/Warning findings. Ten focused repetitions and two
  focused race repetitions passed in addition to the stable full-journal run; the reviewer confirmed exact
  KR/US coexistence and release isolation, legacy-global precedence, migration rollback and the deliberately
  unreachable production official-zero mint.

## Wave 2A (2026-09-25) — HEAD re-settlement VERIFY

- Voices: Opus implementation teammate (this record); independent verification by the Manager is pending.
- Production edits: **none**. Added files: one build-tagged test (`internal/execgw/a066_entry_loss_lock_red_test.go`),
  evidence under `analysis/`. Function Logic Map: `not-applicable` for this lot's code — no existing function body
  was edited; the test file is a new leaf under a build tag.

### Wave 1E re-measurement

GREEN at HEAD after the a112 merge (schema v32): four packages `go test -count=1` rc 0 (journal 495.5 s), focused
`-race` rc 0 (125/125, 0 race reports), riskbucket/officialfx `-race` rc 0, vet rc 0. Full numbers in `status.md`.
No a066 test needed updating for the a112 schema moves.

### Evidence findings

- Wave 1C/1D edited four base functions without writing their bundles (`Journal.RecordFill`,
  `riskbucket.ApplyFill`, two fill test helpers). The Wave 1C text "the only existing fill-path body edit is ...
  `RecordFill`" was true, but the bundle it implied never existed. Closed in this lot.
- Two older maps carried positional drift: `TestSchemaTablesAndColumns` B5/B6 and the old `Gateway.submit`
  B24/B25 descriptions did not match the branches at those positions; `activeScopeWhere`/`scopeArgs` listed a
  fall-through return as `B3`. Rewritten from AST source lines.
- `TestRevokedDecisionIsRefusedAtTheLastMoment`, cited by the old submit map for the last-moment fresh-decision
  refusal, never enters that branch (submit B30, 0 executions); per-test coverage shows it is refused at the
  initial `checkReservation` (submit B28 → checkReservation B6). Citations now follow per-test coverage.
- Measured coverage gaps (statement coverage, not per-branch-arm): submit 16/57 bodies unexecuted — including
  **B41, the strategy-plan copy of the a066 last-moment q_final barrier**; RecordFill 19/38 — including a066's
  storage-error exits B32/B35; ApplyFill 19/38 — mostly corrupt-state parse exits, plus B1 (invalid identity/
  quantity refusal) and B38; checkReservation B2 (reservation read error). None is new in this lot; they are
  listed for 6.1 ("RED-to-GREEN evidence for every Branch Test Map row").

### 2.7 RED — `internal/execgw/a066_entry_loss_lock_red_test.go`

- Build tag `a066_red_5_5`; no make target sets it, so `make test`, `make test-seams`, `make lint` are unaffected
  (untagged `go test -run TestA066 ./internal/execgw` → "no tests to run"; `go vet` untagged and with
  `tossos_testseams` rc 0; `go vet -tags a066_red_5_5` rc 0).
- Contract: `TestA066EntryLossLockIsEntryOnlyPerHorizonAndMarket` (a lock refuses only its own market×horizon
  q_final entry and writes no decision/hold; MEDIUM does not reach SHORT; US does not reach KR — spec
  "Medium lock과 short entry") and `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths` (with KR/US ×
  SHORT/MEDIUM locked and a q_final bucket failure live: KR stop sell, US exit sell, reduce-only cancel reach the
  broker once each inside a 5 s deadline, a KR reconcile entry is recorded, the exit's fill is applied with delta
  2, and a fresh exposure-raising decision is still refused before and after the risk-reducing traffic).
- RED log (`go test -count=1 -tags a066_red_5_5 -run 'TestA066' -v ./internal/execgw`, rc 1, 4.4 s; abridged — the four failing subtests each print the same message at line 108):

  ```text
  --- FAIL: TestA066EntryLossLockIsEntryOnlyPerHorizonAndMarket
      --- PASS: .../control:_no_lock_admits_KR_SHORT
      --- PASS: .../control:_no_lock_admits_KR_MEDIUM
      --- FAIL: .../KR_MEDIUM_lock_does_not_reach_KR_SHORT
      --- FAIL: .../KR_SHORT_lock_refuses_KR_SHORT
      --- FAIL: .../KR_MEDIUM_lock_refuses_KR_MEDIUM
      --- FAIL: .../US_locks_do_not_reach_KR
  a066_entry_loss_lock_red_test.go:194: [RED, 5.5 대기] a066 entry loss lock 이 없음: horizon×market 잠금을
      durable 로 활성화하는 seam(activateEntryLossLock)이 연결되지 않았음 — task 5.5 가 구현해야 함
  --- FAIL: TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths
  ```

- Non-vacuity probe (temporary uncommitted file setting the seam to a no-op, deleted after the run): the two
  lock-refusal rows fail with "locked SHORT/MEDIUM entry was admitted … q_final=10"; the control, isolation and
  risk-reducing rows pass. So the refusal rows go red for the right reason, and the risk-reducing test is a
  guard that a real lock must not break (a no-op lock cannot break it, which is expected).
- The probe also caught a fixture error before commit: a US **market** sell is refused by the trading policy
  (`live place supports only a narrow subset of orders`) regardless of any lock, so the US exit uses a limit sell.

### Open design points for 5.5 (not user decisions)

- A q_final decision issued **before** a lock activates and submitted after it: the spec blocks "신규
  EXPOSURE_RAISING decision과 추가 leg"; the conservative reading refuses the submit at Gateway revalidation.
  2.7 does not pin this — 5.5 should add a row for it.
- The build tag must be removed in the 5.5 GREEN commit; a tag nobody runs is the failure mode recorded in memory
  ("tagged tests never ran").

### Environment incident

- The root filesystem reached 100 % during the lot (Go build cache 208 GB, grown 2026-09-22…25). Coverage runs
  failed with `database or disk is full`; so would anyone else's journal tests. I deleted Go build-cache entries
  last modified before 2026-09-24 00:00 (Go refreshes an entry's mtime when it is used, so these were unused for
  over a day), freeing ~122 GB (cache 208 → 86 GB). Cache only; no repository or user file was touched.

### 사용자 결정 대기

- 없음 (this lot needs no human decision; 5.5's relaxation path will).
- **5.5 설계점 확정 (2026-09-25, 사용자 결정 ⑤ = 거절)**: 락 **전**에 나온 q_final 결정이 락 **뒤**에 제출되면
  Gateway 재검증이 제출을 **거절**한다. 근거 — 스펙이 막는 것은 「신규 EXPOSURE_RAISING decision 과 추가 leg」이고
  노출은 제출 시점의 상태다; 락 상태와 제출을 둘 다 보는 자리는 Gateway 뿐; 불변식 6(사이징은 보수 방향만);
  a112 관문이 조정자 Submit 앞에 서는 것과 같은 자리. 5.5 는 이 행을 2.7 의 RED 표에 먼저 더하고 `a066_red_5_5`
  태그를 GREEN 커밋에서 뗀다. 완화 경로(사람 승인·audit 되는 relaxation)는 5.5 로트가 설계를 들고 올 때 사용자에게 묻는다.

## 5.5 lot (2026-09-27) — 결정 ⑤ RED 행 · 스키마 정지점

- Voice: Opus implementation teammate. Production edits: **none**. Function Logic Map: `not-applicable` for this step —
  only the build-tagged test file gained a new test function; no existing function body was edited.
- 결정 ⑤ RED 행: `TestA066DecisionIssuedBeforeLockIsRefusedAtSubmit` (same tag `a066_red_5_5`). A KR SHORT q_final
  decision is issued with no lock, then a lock is activated, then the decision is submitted through a Gateway on the
  same journal. Rows: control (no lock → broker place 1, PASS today), KR MEDIUM lock (must still submit),
  KR SHORT lock (must be refused, `StateNotDispatched`, broker places 0). The reason code is deliberately not pinned —
  a new enum is a contract choice for 5.5 GREEN.
  - RED (`go test -count=1 -tags a066_red_5_5 -run TestA066 -v ./internal/execgw`, rc 1): the two lock rows stop at the
    missing seam; the control row PASSes.
  - Non-vacuity probe (temporary pid-named file setting the seam to a no-op, deleted after the run, rc 1): the refusal
    row fails with `pre-lock decision submitted under its own lock: … state=CONFIRMED places=1`; control and the
    MEDIUM-lock isolation row PASS.
  - `go vet` untagged and `-tags a066_red_5_5` rc 0; gofmt clean.
  - Commit of the test file is held: Manager froze Go commits during the a071 gate sequence (2026-09-27).
- **정지점 — journal schema 변경 필요 (SchemaVersion 32 → 33, main=HEAD=32)**. The seam contract
  `activateEntryLossLock(ctx, *journal.Journal, account, market, horizon, at)` needs durable market×horizon state that
  the Gateway revalidation can read. No existing table can hold it without a DDL change (measured on HEAD `9494e0e6`):
  `risk_bucket_scope_latches` has `CHECK(latch IN ('RISK_OVERAGE','UNKNOWN_ACTUAL_RISK','REPLAY_MISMATCH','ORPHAN_FILL'))`
  and a per-symbol/prospective-generation key with no horizon column; `operating_modes` has
  `CHECK(mode IN ('NORMAL','ENTRY_BLOCKED','HALT_ALL'))` and is account-wide; `risk_bucket_events` is owner-scoped and is
  read by the owner-release cleanliness checks. Per the lot's rule the work stops here and the questions go to the
  Manager; the relaxation flow (user decision pending since ⑤) is part of the same question set.

## 5.5 lot (2026-09-27, continued) — entry side GREEN, relaxation not implemented

Manager rulings applied (2026-09-27): v33 additive migration approved; relaxation (approval flow, API and callers)
stays **unimplemented** until the user answers — only its place is marked in `risk_bucket_entry_loss_lock.go`; dormant
activation API with zero production callers; one rule function called at every enforcement site; reason code
`ENTRY_LOSS_LOCK_ACTIVE` approved with both layers distinct (riskbucket refusal + Gateway reason, mapped by type);
fixture repair option A (fail-closed stays) with cross-change test edits approved.

### What landed (production)

| File | Change |
|---|---|
| `internal/journal/risk_bucket_entry_loss_lock_v33.sql` | v33: `risk_bucket_entry_loss_locks` (append-only; triggers no-update, no-delete, first-cause-wins per account×market×horizon). No existing table touched. |
| `internal/journal/schema.go` | `SchemaVersion` 32 → 33, migration step 33. |
| `internal/journal/risk_bucket_entry_loss_lock.go` | `ActivateEntryLossLock` (tightening only, idempotent, first cause wins); the single rule `refuseEntryUnderLossLock`; `admissionHorizon`/`decisionHorizon`. Relaxation: placeholder comment only. |
| `internal/journal/risk_bucket.go` `CommitRiskBucketAdmission` (new B13) | rule call inside the admission tx, after the existing entry-scope gate. |
| `internal/journal/risk_bucket_issuance.go` `commitFreshRiskBucketAdmissionTx` (new B7) | same rule inside the tx shared by q_final issuance and strategy first leg. |
| `internal/journal/risk_bucket_issuance.go` `RevalidateQFinalAdmission` (new B18, B19) | last check before `return true, nil`: reads the decision's horizon reservation and calls the same rule (user decision ⑤). |
| `internal/riskbucket/types.go` | `RefusalEntryLossLockActive = "ENTRY_LOSS_LOCK_ACTIVE"`. |
| `internal/execgw/reason.go`, `gateway.go` `checkReservation` (new B7), `failclosed.go` `AllReasonCodes`, `testdata/reason_codes.golden` | `ReasonEntryLossLockActive = "entry_loss_lock_active"`, chosen by `errors.Is(err, journal.ErrRiskBucketEntryLossLocked)` after the Guardian-missing check; registered in the enumeration and golden (generator). |

**Three enforcement sites, not two** — the correction the Manager accepted: AST shows two admission transactions
(`CommitRiskBucketAdmission` and `commitFreshRiskBucketAdmissionTx`, the latter shared by q_final issuance and
`strategy_first_leg_atomic.go:174`) plus the Gateway revalidation. One rule function serves all three.

Lower-case spelling receipt: the comparing side is `TestReasonCodeEnumIsStable` against `testdata/reason_codes.golden`
(lowercase snake, subject_state: `operating_mode_blocked`, `flatten_in_progress`); `entry_loss_lock_active` is the
lowercase form of the approved `ENTRY_LOSS_LOCK_ACTIVE`. `strategyPreTransportReason` was not edited: the new code
falls to its default `GATEWAY_POLICY_REFUSED` (record-only classification, `strategy_dispatch_refusal.go:10-12`).

Risk-reducing paths: no function on the stop / emergency-exit / reconcile / fill path calls the rule.
`RevalidateQFinalAdmission` is reached only through `checkReservation`, whose B1 returns before it for every
non-EXPOSURE_RAISING decision. `TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths` (KR+US × SHORT+MEDIUM
locked) PASS untagged.

### Tests

- 2.7 file `internal/execgw/a066_entry_loss_lock_red_test.go`: build tag removed, seam wired to
  `journal.ActivateEntryLossLock`; expectations unchanged except the two approved additions — refusals must carry
  `ErrRiskBucketEntryLossLocked` + `ENTRY_LOSS_LOCK_ACTIVE`, and the ⑤ refusal must carry `entry_loss_lock_active`.
- `internal/journal/risk_bucket_entry_loss_lock_test.go`: 3 sites × {KR/SHORT, US/MEDIUM} × {no lock, own scope, other
  horizon, other market, other account} (30 rows); first cause wins + invalid input + raw UPDATE/DELETE/duplicate
  INSERT refused while another scope is accepted; survives reopen; 8 concurrent writers leave one lock; rule fails
  closed on unknown scope and on a read error; v32→v33 starts with zero locks.
- `internal/execgw/a066_legacy_entry_census_test.go` (see "legacy residual" below).

### Verification (sequential, one run at a time)

| Run | Result |
|---|---|
| `make lint` (gofmt + vet untagged + vet `tossos_testseams`) | rc 0 |
| `make test` (whole repo, before the a098 fix) | 98 ok, 1 FAIL: `TestTheSenderDownReasonIsRegisteredInTheEnumeration` (count pin, fixed below); journal ok 518 s |
| `make test-seams` (same checkout) | same single failure |
| `go test ./internal/execgw/...` untagged and `-tags tossos_testseams` after the a098 fix | rc 0 / rc 0 |
| `go test -race -run 'LossLock\|MigrationV32ToV33\|QFinal\|RiskBucketAdmission' ./internal/journal` | rc 0 |
| `go test -race -run 'TestA066\|QFinal' ./internal/execgw` | rc 0 |

Note: the checkout is shared; a peer session had uncommitted `internal/app/engine` and `internal/strategyrouter` edits
during these runs, so the repo-wide numbers include them.

### Mutation (copy at `<scratch>/mut-<pid>`, no-mutation control GREEN first; `analysis/harness/mutate_5_5.py`)

- run1 (`analysis/mutation-5.5/ledger-run1.tsv`): 26/29 CAUGHT; **three did not reach** — M12 was "caught" by a compile
  error (`horizon` unused), M19/M20 SURVIVED because the journal `-run` regex (`EntryLossLock`) did not select
  `TestRefuseEntryUnderLossLock…`. Harness fixed (compiling M12, regex `LossLock`), recorded in the harness comments.
- run2 (`ledger-run2.tsv`, 36 mutants incl. reason-code layers and census positive controls): 35 CAUGHT, M15
  NOT-APPLIED (its text predated the typed refusal).
- run3 (`ledger-run3.tsv`, sites + M15 + reason layers, failing test names recorded): all CAUGHT, and **each site's
  plumbing mutant fails that site's own rows** — M01–M04 only `…/CommitRiskBucketAdmission/…` (+ reopen test),
  M05–M08 only `…/RecordQFinalDecisionAndReserve/…` + the 2.7 issuance rows, M09–M12 only `…/RevalidateQFinalAdmission/…`
  + the ⑤ submit row; market/horizon constants fail exactly the US/MEDIUM `own scope` + `other market`/`other horizon`
  rows. M30/M31 (Gateway mapping removed / mapped to mismatch) fail the ⑤ row; M34–M36 (production `NewTracer` call,
  `GuardianAdapter` literal, new EXPOSURE_RAISING producer) fail the census.

### Cross-change fixture repair (option A, Manager-approved)

Mechanism: four top-level tests (nine leaf tests; the earlier "five" miscounted) opened a v25 journal (`migrationOverride`) and wrote q_final rows into it with the **current**
writer; v33's lock read then hit a missing table and failed closed — a latent fixture mismatch that v33 exposed.
Fix in `prepareStrategyDispatchLease` (`strategy_dispatch_runtime_test.go`, a072): rows are produced by the current
writer on a current-schema scratch journal and copied into the older journal. Before copying, the helper asserts
(a) the set of tables the writer touched equals the fixture's list and (b) each table's `sqlite_master.sql` is byte-equal
in both journals — so the row shape is the older migration's own DDL, not memory; after copying it asserts the tables
are row-for-row equal.

| Test | What it asserts (unchanged) | Setup before | Setup after |
|---|---|---|---|
| `TestStrategyDispatchLeaseSchemaRequiresExactQFinalAuthorityAndHolds` (4 subtests) | v25 lease triggers accept an exact sealed row and refuse a fabricated decision, a released monetary hold and a substituted authority digest | q_final rows written into the v25 journal by the current writer | same rows produced on a v33 scratch journal, DDL-equality checked, copied |
| `TestStrategyDispatchBrokerOrderIDCannotCrossKRUSWithinAccount` | one broker order ID cannot be bound in KR and US | same | same |
| `TestStrategyDispatchColdRestartDiscoversOldIssuedClaimedAndSubmitting` (3 subtests) | recovery discovery per state is read-only and fenced | same | same |
| `TestMigrationV26AddsPairedFirstLegAuthorityWithoutChangingV25Rows` | v25→v26 preserves rows and appends columns only | same | same |

No assertion line in these tests was edited; the diff is confined to the helper and its new support functions.

Second cross-change edit: `TestTheSenderDownReasonIsRegisteredInTheEnumeration` (a098) pinned the absolute enumeration
length (29 + 1). It now expects `29 + 1 + len(reasonCodesRegisteredAfterA098)` and separately asserts each later code's
membership; "a098 added exactly one" still holds.

| Test | What it asserts (unchanged) | Before | After |
|---|---|---|---|
| `TestTheSenderDownReasonIsRegisteredInTheEnumeration` (a098, Manager ratified 2026-09-27) | the enumeration grew by exactly the codes someone decided to add, and `ReasonAlertSenderDown` is a member | `len == 29 + 1` (absolute; any later legitimate code turned it red) | `len == 29 + 1 + len(reasonCodesRegisteredAfterA098)` + each listed code is a member |

Why this shape: the named list `reasonCodesRegisteredAfterA098` is the **only** path for registering a new legitimate
entry against this frozen count — a later change must name its code there, so an unnamed addition still turns the test
red. Measured on a copy (`<scratch>/a098probe-<pid>`, deleted afterwards): control GREEN; then
`ReasonGuardianRiskBucketMismatch` added to `AllReasonCodes` **and** the golden regenerated with the sanctioned
generator (so `TestReasonCodeEnumIsStable` passes) → `TestTheSenderDownReasonIsRegisteredInTheEnumeration` FAIL
"AllReasonCodes() has 32 codes, want 31". The golden alone does not catch an unnamed addition; this count does.

### Legacy residual — measured, not assumed

Entry decisions without the q_final marker have no horizon and are outside the lock (`RevalidateQFinalAdmission`
returns `(false, nil)`). Measured over non-test sources in `internal/` and `cmd/` (2026-09-27): EXPOSURE_RAISING
decision producers are exactly four — `Issuer.IssueEntry`, `RiskGuardian.IssueEntry` (legacy) and
`RiskGuardian.IssuePrecheckedQFinalEntry`, `RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg` (q_final). Legacy
`RiskGuardian.IssueEntry` is called only from `Tracer.submitEntry` and `RiskGuardian.IssueStrategyEntry`; `NewTracer`
has **0** non-test callers and `strategydispatch.GuardianAdapter` (the only `IssueStrategyEntry` caller) is constructed
**0** times; `Issuer.IssueEntry` has no caller (flatten uses `IssueReduction`). The production strategy path is the
q_final first-leg bridge (`strategy_entry_supervisor.go:322-325`). **Reachable legacy entry paths from production
assembly: 0.** `TestA066LegacyEntryPathsAreUnreachableFromProductionAssembly` re-parses the tree on every run, so the
census fails the moment the sample fills (positive controls M34–M36 CAUGHT).

### Function Logic Map

Pre-edit bundles (HEAD `406e54de`): `commitFreshRiskBucketAdmissionTx`, `Journal.RevalidateQFinalAdmission`,
`AllReasonCodes` (new); `CommitRiskBucketAdmission`, `Gateway.checkReservation` (existing, AST matched). Post-edit:
all five re-extracted with measured rows (`pertest_cover_5_5.sh`, 111 tests; execgw package coverprofile);
`TestSchemaTablesAndColumns` (+2 lines, same 7 branches), `Gateway.submit`, `loadRiskBucketState`,
`TestSchemaIndexes` re-extracted because their files changed (bodies identical). `prepareStrategyDispatchLease` and the
a098 test are not base functions (their files postdate base `23794f86`).

### Open

- Relaxation (task 5.5's "human-approved audited relaxation"): user decision pending; Manager carries the recommended
  shape (AUTO tightens only; relaxation = OPERATOR + approval reference + audit line before commit; tossctl mutating
  command, no console button). 5.5 stays unchecked until then.
- Deployment note: this branch is now **schema v33, main is v32** — the rule "SchemaVersion differs from main → do not
  build the image" is in force from this commit. Production behaviour change before deploy: none — the table starts
  empty on migration (`TestMigrationV32ToV33StartsWithNoEntryLossLock`), `ActivateEntryLossLock` has zero production
  callers, and with no lock the rule admits exactly what it admitted before (control rows of every site).
- `RevalidateQFinalAdmission` B18 (horizon reservation read error) is not executed by any test — a storage-error exit
  after B16/B17 proved all five HELD rows exist.
- Independent adversarial review + gstack review: pending (next step of this lot).

## 5.5 review round 1 (2026-09-27) — three voices, findings and dispositions

Voices: Claude `code-reviewer` (lock semantics), Claude `code-inspector-tester` (tests/fixtures, 17 own mutants on a
`git archive` copy), Codex CLI 0.154.0 `exec --sandbox read-only` (static). Verdict of all three: **no P0** — no q_final
path admits or submits exposure inside a locked scope, and no risk-reducing path waits on or is refused by the lock.

| # | Finding (source) | Sev | Disposition |
|---|---|---|---|
| 1 | Activation stores the raw account; `" acct-7 "` lock never matches the trimmed entry account (Codex, reviewer 1 — probed) | P1 | **Fixed**: `validEntryLossLockScope` rejects surrounding whitespace (no silent normalisation); test rows + mutant M37 |
| 1b | Case variants / unknown account ids are still accepted (reviewer 1) | P1 | **Open → Manager/trigger lot**: canonical account identity belongs to the caller that wires activation (0 callers today); binding activation to a journal-known account would refuse a first-trade lock and is a design choice |
| 2 | First-leg site (production strategy path) untested — a first-leg-only skip survived every test (reviewer 2, R04) | P1 | **Fixed**: `TestEntryLossLockRefusesTheStrategyFirstLegAdmission` (KR/US, own/other horizon/market); M42 (two-file skip) CAUGHT |
| 3 | Strategy path: a lock refusal becomes `ATOMIC_ADMISSION_FAILED` and latches the whole market worker (KR SHORT lock also stops KR MEDIUM; counted as worker failure) (reviewer 1) | P1 | **Open → Manager**: over-blocks in the safe direction; strategy lanes are dormant; the fix is a typed refusal through `strategy_first_leg_admission.go` / `strategy_entry_supervisor.go`, files a peer (a112) session is editing now |
| 4 | Lock committed between the last `checkReservation` and `plan.call` is not seen (Codex P1, reviewer 1 P2) | P1/P2 | **Accepted residual**: closing it needs a DB tx held across the broker call; same window as every other durable-latch recheck in the Gateway. The last-moment check itself is now measured (#5) |
| 5 | No lock test activated the lock after the initial Gateway check (reviewer 1) | P2 | **Fixed**: `TestA066LockActivatedAfterInitialCheckIsRefusedAtTheLastMoment`; M38 (remove last-moment check) CAUGHT |
| 6 | Census holes: method values, `new(T)`, `var x T`, `&Tracer{}`, `SafetyClass` by assignment or string literal, `tools/` not walked, root coverage unasserted (reviewers 1, 2) | P2 | **Fixed**: identifier-level `references(...)` cases, producer matcher covers key-value and assignment with any non-RISK_REDUCING value (four class-copying sites measured and pinned with their source lines), roots `internal,cmd,tools` with per-root file-count assertion; M40, M41, M45, M46, M47, M49 CAUGHT |
| 7 | `assertRaisingRefused` after risk-reducing traffic is refused by bucket mismatch regardless of the lock, so "the lock is not consumed" was unproven (reviewer 2) | P2 | **Fixed**: the test now re-activates all four scopes after the traffic and requires `changed=false` with the original cause |
| 8 | v33 CHECK constraints unpinned (R06/R07) (reviewer 2) | P2 | **Fixed**: `TestEntryLossLockSchemaRefusesValuesTheRuleCannotMatch`; M43, M44 CAUGHT |
| 9 | Returned `ActivatedAt` keeps nanoseconds, stored value is seconds (reviewer 1 — probed) | P2 | **Fixed**: truncated to the stored precision; test compares the returned lock to the re-read one; M39 CAUGHT |
| 10 | Mutation ledgers not bound to a tree (reviewer 2) | P2 | **Fixed**: harness writes `TREE HEAD <sha> uncommitted tracked: …` as the first ledger line; it now also copies `tools/`. Runs 1–4 predate this: they ran on the working tree that carried a peer's uncommitted `internal/app/engine` and `internal/strategyrouter` edits (not in any mutated file; the census reads those sources and pinned the same sites) |
| 11 | Replay branches return before the lock check (`risk_bucket_issuance.go` replay, first-leg replay, `CommitRiskBucketAdmission` replay) (reviewers 1, 2) | P2 | **Accepted, recorded**: a replay creates no new authority, and the Gateway revalidation refuses the submit (single layer, measured by the ⑤ rows). `CommitRiskBucketAdmission` has 0 non-test callers |
| 12 | Gateway typed mapping has only a positive test; an unknown-scope/read-error Blocked is reported as `guardian_risk_bucket_mismatch`; on the strategy path the final in-`call()` refusal is wrapped as `strategy_dispatch_fenced` (reviewers 1, 2, Codex) | P2 | **Accepted, recorded**: all refuse; only the operator-visible reason differs. The fenced wrapping is pre-existing a112/a072 code |
| 13 | `first_cause_wins` allows one row per scope forever — after a future relaxation, re-locking the same scope would abort (reviewer 1) | P2 | **Carried to the relaxation decision**: the relaxation migration must replace this trigger (DROP/CREATE TRIGGER, as v23 did) with "one *open* lock per scope" |
| 14 | `IssueQFinalEntry` returns the lock error raw, not as an `IssueRefusal` (reviewer 1) | P2 | Recorded; typed by `errors.Is` + `riskbucket.IsRefusal`, which the tests pin |
| 15 | Equivalent / clean mutants: lock check moved after owner insert (tx rollback), Revalidate check order, `ToUpper` market (reviewer 2 R01, R03, R16) | — | Equivalent; noted |

Re-verification after the fixes (sequential, one run at a time) and mutation run 6 (all mutants, tree-bound):
see the next block.


### Re-verification after the review fixes (sequential)

| Run | Result |
|---|---|
| `make lint` | rc 0 |
| `go test ./internal/execgw/...` untagged / `-tags tossos_testseams` | rc 0 / rc 0 |
| `go test -timeout 60m ./internal/journal/...` | rc 0 (697 s) |
| `-race` journal `LossLock\|MigrationV32ToV33\|QFinal\|RiskBucketAdmission\|FirstLeg` / execgw `TestA066\|QFinal` | rc 0 / rc 0 |
| mutation run 6 (`analysis/mutation-5.5/ledger-run6.tsv`, all 48 mutants M01–M47, M49; no-mutation control GREEN) | **48/48 CAUGHT** |

run 6 tree line: `HEAD fe3db326` plus uncommitted tracked files — the peer a112 edits in `internal/app/engine`,
`internal/strategyrouter` and this lot's own uncommitted fixes (committed right after as the review-fix commit).
The peer's edits were committed during/after the run as `272b715c` (a112 8.7.2, 22:07); every other commit since
`0004536c` is docs-only. None of those Go files is a mutated file or an import of `internal/journal`/`internal/execgw`
tests; the census reads their source and pinned the same sites. `check_analysis.py --change a066-…`: 0 a066-bundle
findings; the 371 `missing evidence` lines are the stacked-window set, none of them a function this lot edited.

### Manager rulings on the open review items (2026-09-27)

- **(a) account identity — canonical form checked at the API boundary; existence left to the trigger lot.** The
  comparing side is the contract: `DecisionRequest.build` and `ReserveRequest.build` apply `strings.TrimSpace` to
  `AccountRef` and keep its case, and admission and Gateway revalidation compare that value exactly (SQL `=`). The
  lock uses the same form: surrounding whitespace is rejected (`7aa158bb`). A case-variant id is a **different
  account** in the journal, so it is not folded. `TestEntryLossLockAccountFormIsTheDecisionBuildersForm` pins both
  the builder form and the exact-match behaviour (`ACCT-1`/`Acct-1` locks do not refuse `acct-1`; an `acct-1` lock
  does). **Named residual:** "is this account real" is for the caller that wires activation (the trigger lot), which
  has the live-account context.
- **(b) named residual:** on the strategy path a lock refusal is flattened to `ATOMIC_ADMISSION_FAILED`
  (`internal/app/engine/strategy_first_leg_admission.go:88-90`), then goes through
  `strategy_dispatch_cycle.go` `dispatch` → supervisor → `latchMarket` (`strategy_entry_supervisor.go`), which latches the
  whole market worker. That over-blocks in the safe direction (a KR SHORT lock also stops KR MEDIUM), and the lanes are
  dormant. The fix is a typed refusal through those a112 files. It belongs to whichever lot wires activation (5.6/5.7 or
  later); the Manager relays this to the a112 HANDOFF.
- **(c) accepted:** the window between the last `checkReservation` and the broker call is the same one every durable
  latch recheck in the Gateway leaves open.

### Relaxation (pending user decision) — constraint to carry

The v33 trigger `risk_bucket_entry_loss_lock_first_cause_wins` allows **one row per account×market×horizon, ever**.
Once a relaxation record exists, locking the same scope again would hit `RAISE(ABORT)`. The relaxation migration must
therefore replace this trigger with "at most one **open** lock per scope" (DROP/CREATE TRIGGER, as the v23 step did).
The same migration must also switch `activeEntryLossLock` to "no relaxation record". This constraint is merged into
the user question about the relaxation flow.

## 5.6.1 (2026-09-27/28) — shared-bucket consistency repair exposed by the 5.6 measurement

Manager rulings: v34 additive (committed v33 untouched); F2 keys the reservation-policy record by the digest the
writers already compute; F1 option (a) — recompute ledger usage inside the admission transaction with the production
reader's own function and refuse an understated snapshot as a retryable stale; refusal code `BUCKET_USAGE_STALE`
(golden `SUBJECT_STATE` form; distinct from the time-expiry `STALE_BUCKET`), sentinel wraps `ErrSnapshotStale`, no new
Gateway reason (the refusal happens at issuance). Attribution: 5.6.1 under 5.6, not a 5.2 reopening.

### RED → GREEN, one defect at a time (the masking made visible)

Contract tests: `internal/journal/a066_integration_5_6_test.go` (promoted from the measurement file; the adaptation
record is at the top of the file). Runs in an isolated copy (`analysis/harness/test_in_copy.sh`) because a peer
lot's untracked RED file in `internal/journal` does not compile in the shared worktree.

| Step | F2 contract (`…SecondAdmissionAtAnotherPriceIsCappedNotCollided` KR/US) | F1 contract (`…StaleSnapshotIsRefusedAtAnyPrice`) |
|---|---|---|
| 0 — before any fix | FAIL: `immutable policy collision` | KR same price: FAIL, **admitted, strategy usage 100 > 80**; KR/US other price: FAIL, refused for the wrong reason (`immutable policy collision`) — F2 masks F1 |
| 1 — F2 fixed (v34) | PASS (q_final 5, usage 80) | all three FAIL, **admitted, usage 100/110/110 > 80** — the mask is gone and F1 is live at every price |
| 2 — F1 fixed | PASS | PASS (refused `BUCKET_USAGE_STALE`, nothing written, usage stays 50) |

### What changed

| File | Change |
|---|---|
| `internal/journal/risk_bucket_policy_records_v34.sql`, `schema.go` | v34: `risk_bucket_policy_records` (PK key + record digest, immutable); `risk_bucket_reservations.policy_record_digest` (nullable; required and must name a record of the same key for v34 inserts; immutable once set). v22–v33 rows keep NULL and keep reading `risk_bucket_policies`. No existing table's rows touched. |
| `risk_bucket_policy_records.go` | `storeRiskBucketPolicyRecord`: same digest as before (`riskBucketRecordDigest{Key, PolicyEvidence, ReservePolicy}` — read from both writers), parent row INSERT OR IGNORE (FK parent of snapshots), record stored and re-read. The key-only collision check (the F2 defect) is gone. |
| `CommitRiskBucketAdmission`, `insertFreshRiskBucketReservations` | call the helper; reservations carry `policy_record_digest`. |
| `loadRiskBucketOrderAuthority` | v34 rows join their own record; NULL rows keep the key join. |
| `internal/riskbucket/production_snapshot_authority.go` | `ReadJournalBucketUsage` (read + sum, the only computation) over a `UsageQueryer` (*sql.DB or *sql.Tx); `aggregateProductionRiskUsage` reports `Latched` instead of erroring; `loadProductionRiskEntries` refuses `Latched` with the unchanged message — production behaviour unchanged. |
| `risk_bucket_usage.go` | `refuseStaleBucketUsage`: per bucket, ledger held+filled via `ReadJournalBucketUsage` inside the admission tx; claimed < ledger → `RefusalError{BUCKET_USAGE_STALE}` wrapping `ErrRiskBucketUsageStale` (wraps `ErrSnapshotStale`); unreadable ledger → non-retryable `ErrRiskBucketSnapshotMismatch`. Latch is not judged here (it stays with `ensureRiskBucketEntryScopeClean`). Called in both admission transactions after the owner/scale-in checks and before any write. |

"Same sum as the production reader" is proven by the call graph, not prose: `TestA066LedgerUsageHasOneComputation`
pins that `readProductionRiskUsage`/`aggregateProductionRiskUsage` are called only by `ReadJournalBucketUsage`, which
is called only by `loadProductionRiskEntries` and `refuseStaleBucketUsage`, which is called only by the two admission
transactions.

Recollection termination (Manager condition): `TestA066StaleUsageRecollectionTerminates` — a stale refusal followed by
a ledger-true recollection is admitted at q_final 6 after exactly 2 collections; a ledger that stays ahead ends after
exactly `MaxAttempts` collections with `ErrRecollectionExhausted` wrapping the stale refusal. The bounds are
`recollectLoop` (`reservations.go`): `for attempt := 1; attempt <= policy.MaxAttempts` and `now.After(deadline)` on
`policy.Budget`.

### Test fixtures that encoded the defects (a066 and cross-change)

| Test | Before | After (assertion target unchanged unless stated) |
|---|---|---|
| `TestRiskBucketAdmissionRejectsImmutablePolicyAndSnapshotCollision` (a066) | asserted the F2 defect: a different policy record under the same key must be refused | **contract changed by 5.6.1**: the second record is admitted and each admission binds its own record; the snapshot-ID collision half is unchanged |
| `TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser` (a072) | KR limit 100 and US reservation 73500 were both admitted into the shared horizon/strategy/sector buckets — true only on top of F1 | shared limit set to 1000000 (the US fixture's value) and each recollection reads ledger usage; still asserts serialisation + both markets admitted |
| `prepareStrategyDispatchLease` (a072) | DDL-equality between v25 and current tables | column-level check: old columns ⊆ new, the only differences are those declared by migrations after v25 (`schemaAdditionsAfterV25`); rows copied by the old journal's columns; post-copy equality over those columns |
| `TestRiskBucketActualAndReleaseRequireExactOwnerDecisionScopeForCollidingOrderID`, lease fixture | stacked admissions claimed usage 0 | snapshot usage read from the ledger (`refreshSnapshotUsageFromLedger`, same function) |
| `TestRiskBucketAmbiguousSidecar…` | corrupt-duplicate copy omitted the new column | copies `policy_record_digest` too |
| `TestStrategyDispatchCycleRunsKRUSConcurrentlyUnderOneCentralOwner` (a112, `internal/app/engine/strategy_dispatch_cycle_test.go`) | both markets `CONFIRMED` in one wave on the shared SHORT horizon bucket — true only on top of F1 | see the next section (Manager ruling 2026-09-28, cross-change edit) |

### Cross-change edit — a112 paired dispatch cycle test (Manager ruling 2026-09-28)

Cause. The test failed **deterministically** at `2b36ae44` (3/3 with `-tags tossos_testseams`; the loser alternated
KR/US): `ATOMIC_ADMISSION_FAILED: … BUCKET_USAGE_STALE (horizon): horizon bucket "SHORT" snapshot claims 0 used, the
ledger holds 882006`. Its fixture (`pairedStrategyDispatchCycleFixture` → `newStrategyRiskLoaderFixture`) gives both
continuation lanes horizon SHORT (from the lane lineage), the same strategy and the same sector, on one account. Its
risk snapshot source is a static fixture journal, so both dispatches claim usage 0. Under the 5.6.1 semantics the second
admission of one wave on a shared bucket is refused as stale and admitted on the next wave. The fixture encoded the old
(F1) meaning. Disabling `refuseStaleBucketUsage` in a copy makes both markets `CONFIRMED` again (mutant below), which
confirms the cause.

Why not "different horizons per market": the horizon comes from the lane lineage/descriptor, and strategy and sector
are shared too. Splitting them would change the fixture's lanes, not its buckets. The test was rewritten to the new
contract instead (Manager option 2). "Admitted on the next wave" cannot be expressed here because the snapshot source
is static. It is measured at journal level by `TestA066KRUSConcurrentContract` and
`TestFirstLegAtomicAdmissionSameAccountKRUSRecollectsTheSerializedLoser`.

| Old assertion | New assertion | Preservation |
|---|---|---|
| both dispatches start together (`start` channel, no pre-ordering) | unchanged | concurrency kept verbatim |
| every result `err == nil && State == CONFIRMED` | exactly one `CONFIRMED`; the other refused with `engine: first-leg admission ATOMIC_ADMISSION_FAILED` **and** `BUCKET_USAGE_STALE`; any other outcome fails | **replaced** — "both confirmed" is exactly what approved 5.6.1 forbids on a shared bucket. It is the one assertion changed in meaning. |
| `seen[KR] && seen[US]` | `admitted.market != refused.market`, one each | both markets' results must still arrive |
| `len(calls) == 2` | `len(calls) == 1` and its market is the admitted one | stronger: the refused market must not reach the Gateway |
| the two leases have equal `OwnerEpoch`/`FencingToken` | the admitted lease equals `cycle.owner.owner`, and that owner's `Epoch == 1` (acquired once); the refused error carries the `first-leg admission` prefix, which `dispatch` emits only after `dispatchOwner` (`strategy_dispatch_cycle.go:147` before `admit` at `:151`) | single central owner: both dispatches passed the one cached owner, and it was acquired once. The refused side holds no lease, so this equality replaces the lease-to-lease one. |

No turn-taking: the old test had no timing assertion beyond the shared start; none was added or removed.

Verification (isolated copy, HEAD `2b36ae44` plus this edit): the test alone `-count=10 -tags tossos_testseams` PASS;
`-count=50 -race` 50/50 PASS; full `internal/app/engine` with `tossos_testseams` PASS (873), `go vet` clean. Mutant
"`refuseStaleBucketUsage` returns nil" → 3/3 FAIL (both markets `CONFIRMED`, `refused=[]`). The unmutated control was
GREEN first.

### Named residual — production snapshot reader schema pin (Manager ruling 2026-09-28: not raised in this lot)

- Coordinate: `productionRiskJournalSchema = 27`, `internal/riskbucket/production_snapshot_authority.go:33`, checked by
  `loadProductionRiskEntries` (`PRAGMA user_version` must equal it exactly).
- Consequence: **until this pin is changed, production q_final strategy entry cannot open.** On every current journal
  (v32 on main, v34 on this branch) the production bucket snapshot is refused, so the strategy path never reaches
  admission. It is fail-closed, it predates this lot, and no function is lost today because the lanes are dormant.
- Why not raised here: raising it is a verdict change that opens a fail-closed production state, and "exact version"
  needs a designed replacement.
- Replacement design candidate: the column-level check used in the a072 fixture repair. Assert that every column the
  reader selects exists (`pragma_table_info`) and that the only differences from the reader's verified schema are
  declared additive migrations. Do not pin one number.
- Owner: the lot that wires the activation caller (the same place as real-account validation and the trigger). This
  also goes into the user report.

### Record — what worked in this lot (Manager, 2026-09-28)

- The masking was undone step by step and measured (step 0 → 1 → 2 above). Each defect was shown RED on its own
  before its fix, not only the first failure.
- A call-graph test proves that the two sums are identical (`TestA066LedgerUsageHasOneComputation`). This is the
  structure itself, not a claim in prose.
- Round 1 mutation left three SURVIVED (CRA-site plumbing, held ignored, ledger read error admitted), and N14 was
  "caught" only by a compile error. Test additions (`TestA066StaleUsageRuleEdgesAtBothAdmissionSites`) plus a compiling
  N14 closed them. Round 2: 20/20 CAUGHT (`analysis/mutation-5.6.1/ledger-run1.tsv`, `ledger-run2.tsv`).

Deployment note: this branch is now **schema v34**, main is v32. The rule "SchemaVersion differs from main → do not
build the image" still applies.

### 5.6.1 review round 1 (2026-09-28) — Codex, Claude semantics, Claude tests

| # | Finding (source) | Sev | Disposition |
|---|---|---|---|
| 1 | Shared-bucket usage latched on another symbol (UNKNOWN_ACTUAL_RISK / RISK_OVERAGE) was ignored by the admission comparison, so B could add exposure on top of a lower-bound `filled` (Codex) | P0 | **Fixed**: `refuseStaleBucketUsage` blocks entry (`ErrRiskBucketEntryBlocked`, not retryable) when `usage.Latched`. This matches design D5 (a latch blocks new exposure in every applicable bucket) and the production reader. Test `TestA066LatchedUsageInASharedBucketBlocksNewEntry`; mutant N21 CAUGHT. With the block in place the full journal suite passes (827 s), so no existing test depended on the old behaviour |
| 2 | `BUCKET_USAGE_STALE` was retryable, but production collect keeps the bucket snapshot fixed for the whole wave and re-reads only the reservation version. A retry therefore replays the same snapshot, and the second market in a shared bucket always loses after three wasted transactions. Two tests had modelled a re-collection that production does not do (Claude semantics — measured) | P1 | **Manager (a) — made non-retryable.** The sentinel no longer wraps `ErrSnapshotStale`, and its message now reads "refused for this cycle and re-evaluated on the next snapshot wave". This also removes the P2 "broker snapshot is older" wording. The name `BUCKET_USAGE_STALE` stays accurate: the snapshot is behind the ledger. The two tests now use production-shaped collection: the snapshot is fixed and the version re-read; the refusal comes after one collection; the next wave collects from the ledger and admits. Mutant N08 (make it retryable again) CAUGHT |
| 3 | No index on `(account_ref, bucket_dimension, bucket_value)`: five full scans inside BEGIN IMMEDIATE (Codex P1, Claude P2) | P1/P2 | **Named residual** (Manager): the path is dormant, so traffic today is 0. The activation-wiring lot decides whether to add a v35 index after measuring lock length inside BEGIN IMMEDIATE. Coordinate: `readProductionRiskUsage` query, `internal/riskbucket/production_snapshot_authority.go`; called five times per admission. Possible interaction: a124 assumes a journal transaction takes at most 16 ms, and this scan grows with history |
| 4 | Production reader takes `PolicyRecordDigest` from the key-level parent row, so a v34 row binds the first entry's record (Claude semantics) | P2 | **Merged into the schema-pin residual** (Manager): the fix needs v34 columns in the reader SQL, which conflicts with the v27 pin and the reader's test DDL. The sum is unaffected |
| 5 | "ledger-true claim is admitted" also accepted a cap refusal; `t.Fatalf` was called from collect goroutines (Claude semantics) | P2 | **Fixed**: the test now requires admission; collect uses a non-fatal helper |

Design sentence (Manager 2026-09-28): **within one snapshot wave, the second market entering a shared bucket is refused
in that wave and admitted on the next wave**. This delays entries by one cycle, fails closed, and never applies to
stops or exits.

### 5.7 (non-shared-bucket part)

`internal/journal/a066_integration_5_7_test.go`:
- `TestA066PartialFillCrashThenRetryCommitsEverythingExactlyOnce`: a crash inside the fill transaction leaves nothing (no fill row, no allocation, no event, no latch, no watermark, position unchanged). The retry then commits once (+1 fill, +5 allocations, watermark 4, position 4, HELD 30 / filled 20 with UNKNOWN latched). A redelivery writes nothing and leaves the state digest unchanged. Actual evidence completes `filled=max(20, 48)` once, and repeating it is a duplicate.
- `TestA066LateFillOverageLatchesEveryBucketOnceAndTheExitStaysOpen`: a predecessor-late fill after replacement and a successor-late fill after cancel latch RISK_OVERAGE on all 5 buckets. Redelivering both writes nothing. A SELL fill on the same account and symbol is recorded under the latch without touching bucket accounting, and a new entry is blocked.

Claude tests voice (independent; `git archive b8211926` copy; 25 of its own mutants):

| # | Finding | Sev | Disposition |
|---|---|---|---|
| T1 | The v34 required-record trigger had no behavioural test. `migration_v34_test`'s "no record" INSERT was refused by `UNIQUE(decision_id, bucket_dimension)`, not by the trigger. Ledger N17 "CAUGHT" was an SQL syntax error: unbalanced parenthesis, migration 34 failed. With the trigger disabled correctly, all tests stayed GREEN. Mutants that dropped key columns from the trigger (R16–R19) and the NULL→value-only immutability variant (R20) survived | P1 | **Fixed**: a fresh decision/reservation per row; six rows (no record, unknown digest, another dimension's record, another value, another policy version, own record), each refusal asserted by the trigger's own text; binding change refused for a v34 row (value→value) and a legacy row (NULL→value); a fill-accounting update is still accepted. N17 rebalanced. Added N17b/c/d (trigger ignores value, policy version, whole key) and N18b (guards only NULL→value). **The earlier "20/20 CAUGHT" was overstated: N17 did not reach its target** |
| T2 | Fixtures understated all four shared buckets at once, so a check on any single bucket refused them. Skipping sector (R07) or market (R08) survived | P1 | **Fixed**: `TestA066StaleUsageIsJudgedPerDimension` understates one dimension at a time (filled 49 vs ledger 50) at both admission sites, asserts that the refusal names that dimension, and adds an all-ledger-true case where the usage is filled only (held 0). Added N23 (first bucket only), N24 (sector skipped), N25 (market skipped) |
| T3 | Claim side ignoring `filled` survived (R05) | P2 | **Fixed** by T2's filled-only cases; added N22 |
| T4 | "ledger-true claim is admitted" had slack | P2 | Already fixed; now also asserts q_final 6 |
| T5 | The snapshot-collision half asserted only the shared sentinel | P2 | **Fixed**: also asserts the "immutable snapshot collision" text |
| T6 | Latched rows left out of the admission sum (R24) survived | P2 | Superseded: since the P0 fix, latched shared usage blocks entry before the sum is compared (N21) |
| T7 | Mutation ledgers ran on HEAD plus a peer's uncommitted files; N06 was mislabelled | P2 | Recorded (each ledger's first line names the tree). N06 renamed "held claim inflated tenfold" |
| T8 | Step 0 of the RED→GREEN sequence cannot be reproduced from the committed test (the sentinels did not exist at `b8211926^`) | P2 | Recorded. The step logs from the measurement run are committed as `analysis/mutation-5.6.1/red-step{0,1,2}.log` |
| T9 | a072 serialized-loser never asserted a recollection happened; the lease fixture compares column names only | P2 | The serialized-loser test was rewritten in production shape (one loser per wave, admitted on the next wave). The lease fixture's column-subset check is accepted: inserts into the v25 target still enforce v25 constraints, and the post-copy dump is compared column by column |

Final verification (2026-09-28, sequential, isolated copy): `internal/journal`, `internal/riskbucket`, `internal/execgw`
untagged rc 0 (journal 465 s), `tossos_testseams` rc 0, focused `-race` rc 0. Mutation run 3
(`analysis/mutation-5.6.1/ledger-run3.tsv`, tree `HEAD 56f107cc` + this lot's uncommitted files): **29/29 CAUGHT**,
N17 now refused by the trigger itself (the failing test is `TestMigrationV33ToV34KeepsLegacyReservationsReadable`,
not a migration SQL error).

## 5.6 — KR/US concurrent contract (2026-09-28)

`internal/journal/a066_integration_5_6_contract_test.go` takes the task 5.6 sentence as its contract. No operating toggle is used; the tests exercise the journal admission authority only.

- `TestA066KRUSConcurrentContract` builds one flow in one account:
  - Another account's entry sits in the same bucket values and must never show up in this account's buckets.
  - Wave 1: KR and US arrive concurrently with the same wave-start snapshot. Exactly one is admitted and the other is refused `BUCKET_USAGE_STALE`. Wave 2 admits the loser.
  - US uses a different horizon and sector, so **strategy is the only bucket the two markets share**.
  - **Independent market buckets**: KR 50, US 50.
  - **Shared strategy cap**: strategy holds 100.
  - **One symbol, one owner**: a rival lane on KR 005930 or US AAPL gets an owner conflict and writes no decision. There is one active owner per market.
  - **Monetary scale-in aggregation**: the same owner's second decision reuses the owner and is capped by the shared strategy headroom (120 − 100 → q 4). Usage becomes symbol 70, KR 70, US 50, strategy 120. After that, new KR and US entries are cap-refused.
- `TestA066KRUSFailureIsolation` covers **failure isolation**. Both markets use the same symbol string and horizon, so only the market scope separates them. Each KR-only failure blocks KR and leaves US open:
  - a KR loss lock;
  - a KR-scoped reconcile;
  - an exhausted KR market bucket.

  The contrast case is also pinned: a latch in a shared bucket blocks both markets (design D5), so it is not isolated.
- Contract mutants (`mutate_5_6_1.py --set 5.6`, ledger `analysis/mutation-5.6/ledger-contract.tsv`), each making one part of the sentence false: usage summed over every bucket value (C01) or every account (C02), owner conflict ignored (C03), reconcile scope ignoring market (C04), loss lock ignoring market (C05), stale check skipping strategy (C06), latched shared usage ignored (C07). **7/7 CAUGHT.**
  - Two earlier drafts let mutants survive, and the fixture was fixed each time. C02 survived because only one account was used. C06 survived while horizon and sector were shared (they caught what strategy should). C05 survived while the US isolation entry used a different horizon.


## 5.7 shared part — late/actual fill overage across owners of a shared bucket (2026-09-28)

Manager ruling: repair in 5.7 scope. The latch goes on the owner whose fill caused the overage; other owners are
already blocked by the single 5.6.1 rule (latched shared usage blocks entry). No second copy of the verdict is written.

**Measured defect** (`analysis/mutation-5.6.1/shared-overage-measure.log`, before the repair): owners A (US AAPL) and B
(US MSFT) each fill 50 of the shared horizon, US-market, strategy and sector buckets (limit 100). A's actual fill price
then raises A to 78, so the bucket sum is 128 > 100. **Zero** RISK_OVERAGE latches were set. The cause: fill accounting
(`loadRiskBucketFillTransition` → `riskbucket.recomputeOverageLatches`) compared only the owner's own usage with the
snapshot limit, while design D5 and the 5.6.1 admission use the bucket sum.

**Repair**
- `FillState.SharedUsedMinor` holds, per bucket, the other entries' ledger usage. It is not persisted and not part of
  the state digest.
- `loadRiskBucketFillTransition` fills it through `riskBucketSharedUsage`: ledger total via
  `riskbucket.ReadJournalBucketUsage` (the same function as the admission check and the production reader) minus this
  owner's sum.
- `recomputeOverageLatches` adds it to `used`.
- An invalid ledger row (`riskbucket.ErrJournalUsageInvalid`, a new typed sentinel with the same message text)
  becomes a semantic error: the fill is kept, and REPLAY_MISMATCH plus FILL_UNACCOUNTED are recorded. A read failure
  stays a storage error, handled as before.
- The call-graph test now pins the third caller: `riskBucketSharedUsage` ← `loadRiskBucketFillTransition`.

**Contract** (`internal/journal/a066_integration_5_7_shared_test.go`, promoted from the measurement):
- After A's actual fill, A's horizon, market, strategy and sector reservations carry overage 28 and RISK_OVERAGE.
  A's own symbol bucket (78 of 100) has no overage. The A owner is latched, and B's reservations carry no copy.
- A new entry C, even with a ledger-true snapshot, is blocked by the 5.6.1 shared-latch rule.
- **Falsification scenario (Manager):** B's order is cancelled and its hold released, so the shared sum drops to 78,
  below the limit. A's latch and overage stay, and a new entry D is still blocked. B shrinking does not reopen entry.
- An unreadable shared row keeps the fill and watermark and latches REPLAY_MISMATCH / FILL_UNACCOUNTED.

**Mutation** (`mutate_5_6_1.py --set 5.7`): shared usage not added (S01), loader skips it (S02), owner counted twice
(S03), clone drops the map (S04), invalid usage counted as zero (S05), invalid usage treated as a storage error (S06).
All **6/6 CAUGHT**.
- The first S05 draft targeted the storage-error return, which the invalid-row test never reaches, so it SURVIVED.
  It was re-targeted at the invalid-usage branch, and S06 was added for the opposite misclassification.

**Release participation — cited from existing code, not designed here.** The new latch is the owner's existing
`risk_overage_latched` flag and the reservations' `risk_overage_latched`. `releaseRiskBucketOwner`
(`internal/journal/risk_bucket_owner.go:807`) refuses release while `"owner_latch"` counts a latched owner
(`risk_overage_latched=1 OR unknown_actual_latched=1`, line 923). **No code anywhere clears RISK_OVERAGE**: grep finds
no write of `risk_overage_latched=0` and no reset of `LatchRiskOverage`. So an owner that ever records overage, the
single-owner case that predates this lot included, can never be released under current code. Per the Manager's
condition, the release/clearing design stops here and is reported as a named residual. It belongs with the relaxation
decision: human-approved and audited clearing, like the entry loss lock.

**Relaxation question — merged (Manager 2026-09-28).** The pending user decision now covers two items with the same
shape: human-approved, audited clearing for (1) the entry loss lock and (2) the RISK_OVERAGE latch, which no code in the
repository clears today (`releaseRiskBucketOwner` refuses release while `owner_latch` is set). Both clearing flows are
unimplemented until the user answers.

5.7 shared-part verification (sequential, isolated copy): `internal/journal`, `internal/riskbucket` and
`internal/execgw` pass untagged (journal 460 s), the `tossos_testseams` runs of riskbucket and execgw pass, and
`-race` passes on the journal focus set and on riskbucket. One regression was caught along the way: three riskbucket
crash-pure tests (`TestApplyFill*CrashPure`) failed because `cloneFillState` turned a nil `SharedUsedMinor` into an
empty map. Fixed so nil stays nil. Mutation ledger: `analysis/mutation-5.7/ledger-shared.tsv` (6/6 CAUGHT). FLM:
post-edit tables for `recomputeOverageLatches`, `cloneFillState` and `loadRiskBucketFillTransition`, with pre-edit
tables in `analysis/pre-edit/5.7/`. Four bundles whose files shifted were re-extracted with the same branch shapes.

## 6.3 measurement — dormant, no loosened limit, no toggle, unresolved US FX (2026-09-28, partial)

Measured at `d78f3f4a` (`git grep`, `git show -U0`). **6.3 is not checked yet.** The removals from the 2026-08-04
waves listed below are described but have not been re-reviewed in this lot. 6.5's independent review must confirm
them.

**Dormant by default.**
- `ActivateEntryLossLock` has 0 non-test callers. The only hit is its definition
  (`internal/journal/risk_bucket_entry_loss_lock.go:72`).
- The q_final entry points in non-test code are reached only through the Guardian's own chain:
  `PrecheckQFinalEntry` (`riskguardian_qfinal.go:307`, `riskguardian_first_leg.go:50`) and
  `IssuePrecheckedQFinalEntry` (`riskguardian_qfinal.go:311`).
- `CommitRiskBucketAdmission` has no non-test caller.
- The strategy first-leg path stays closed in production: the snapshot reader's `productionRiskJournalSchema = 27`
  pin refuses every current journal (named residual above).

**No toggle flipped.** The a066 lot's production commits since the base (`0004536c`, `7aa158bb`, `b8211926`,
`bb44e7af`, `2b36ae44`) touch 13 non-test Go files: execgw failclosed/gateway/reason, journal risk_bucket*/schema,
and riskbucket fill/production_snapshot_authority/types. None is a config or toggle file. Their added lines contain
no toggle, enabled, automation or live identifier.

**No limit loosened — deletions in those five commits, classified.**

| Commit | Removed | Classification |
|---|---|---|
| `0004536c`, `b8211926` | `SchemaVersion` 32→33→34 | additive migrations |
| `7aa158bb` | `TrimSpace(account) != ""` | stricter: a padded account is now refused |
| `b8211926` | "immutable policy collision" refusal (two sites) | **the one removed refusal.** It was replaced by per-admission policy record binding (F2, Manager-approved 5.6.1 contract). Caps and snapshot-ID collision checks are unchanged. |
| `b8211926`, `2b36ae44` | production usage reader rewritten (`ReadJournalBucketUsage`) | still refuses invalid rows (`ErrJournalUsageInvalid`) and latched rows (`Latched` → refused in `loadProductionRiskEntries`) |
| `bb44e7af` | `ErrRiskBucketUsageStale` no longer wraps `ErrSnapshotStale` | stricter: not retried in the same cycle |

No numeric cap or limit constant was deleted or changed.

**Earlier a066 waves (2026-08-04) — removals found, not re-reviewed here.**
- `4a364caf` removed two interim refusals, "scale-in while risk order accounting is active" and "multi-decision owner
  fill aggregation is not active". The same commit ships multi-decision aggregation, which those refusals stood in for.
- `ff857fae` replaced the US FX `Source != official-fx` check with the sealed FX authority.
- `a37d97f5` removed a comment block in `gateway.go`.
- `9bf1a3a6` added a market scope to reconcile states. `activeEntryScopeWhere` still blocks both markets on a legacy
  NULL-market row. A market-scoped row blocks only its own market; release is exact.
- `ff857fae` moved the `official/client.go` HTTP client to an explicit transport with the same `defaultTimeout`.

**Unresolved US FX → q_final 0.** Existing tests cover this.
- `TestQFinalAccountBaseFXRefusalMatrixKRUS`: missing/stale/wrong-pair/cross-market for both markets. Each is refused
  with zero recollections and no decision row.
- `TestQFinalPrecheckUSUSDWithKRWGuardianFailsCurrencyUnresolvedAndCollectsNothing`
- `TestQFinalPrecheckRejectsCrossedMarketCurrencyPairs`

These tests have not been re-run in this lot yet; that is part of 6.1.

`openspec validate a066-add-multi-horizon-risk-buckets --strict --no-interactive`: rc 0 at `d78f3f4a` (part of 6.4).

### F2 framing check — is policy immutability actually carried over? (2026-09-28, measured; Manager decision pending)

The Manager accepted removing "immutable policy collision" on one condition. A test must show that when a policy's
**content** really changes under the same `policy_version`, the admission is still refused. **No such test exists,
and the behavior is the opposite.** Probe (`analysis/harness/probe_policy_version_digest_test.go.txt`, not
committed as Go):

| Tree | Second admission: same keys and `policy-v1`, every policy digest changed (`NewPolicyProvenance` resealed, `ref.PolicyDigest` matched) | Distinct `policy_digest` under `policy-v1` (horizon) |
|---|---|---|
| `b8211926^` (pre-F2) | refused: `immutable policy collision` | 0 (`risk_bucket_policy_records` absent) |
| `d78f3f4a` (HEAD code) | **admitted** (`err=<nil>`) | 2 |

Logs: `analysis/mutation-5.6.1/probe-policy-version-{pre-f2,head}.log`.

Why the pre-F2 refusal cannot simply come back, and why the journal cannot tell the two cases apart:
- The production policy digest mixes pricing into policy identity. `production_snapshot_authority.go:391` computes
  `policyDigest = H(ManifestDigest, body.PolicyVersion, dimension, value, StrategyRiskVersion, Price.Digest, FX.Digest,
  Fee.Digest)`, so every entry at a new price has a new "policy digest" under the same version.
- The old collision refusal therefore also fired on legitimate price changes; that was F2. It never isolated a
  content change.
- `CommitRiskBucketAdmission` (`risk_bucket.go:308`) checks only that the evidence digest equals `ref.PolicyDigest`
  within one admission. Nothing compares a later admission's policy content with the first one for the same version.
- The limit itself sits in the snapshot (`limit_minor`). No journal check compares it across snapshots of the same
  version, and none did before F2 either.

So "immutable policy per version" is currently guarded only upstream: signed manifest → `ManifestDigest`, and
`StrategyRiskVersion` → `PolicyVersion`. It is not guarded in the journal. Whether one version can bind two manifests
has not been measured.

Options, for the Manager:
- (a) The production reader emits a separate content digest (`H(ManifestDigest, body.PolicyVersion, dimension,
  value, StrategyRiskVersion)`, no pricing). The journal refuses a different content digest under the same key. This
  needs a new column (schema v35), so it stops at the schema rule.
- (b) Record it as a named residual. Content immutability is an upstream (manifest) property, to be proven where the
  version is minted.

**Manager decision (2026-09-28): (b), a named residual.**
- The journal never guaranteed "one version = one policy content", before F2 or after. The old collision refusal
  hashed a digest that includes pricing, so it wrongly refused legitimate price changes and never singled out a
  content change (measured: `probe-policy-version-{pre-f2,head}.log`).
- Content immutability is an upstream property. It is to be proven where the version is minted, along the chain
  signed manifest → `ManifestDigest` → `StrategyRiskVersion`.
- Named residual, not measured: whether one version can be bound to two manifests. The place to verify it is the
  version-minting path.
- Why (a) was rejected: a new content guard in the journal would copy the upstream judgement (with two judgements,
  neither can be disproven), and it needs schema v35.
- The framing "the v34 record binding carries policy immutability with the right identity" is withdrawn, because the
  measurement contradicts it.

## 6.1 / 6.2 verification lot (2026-09-28)

All runs are sequential, in isolated copies (`analysis/harness/test_in_copy.sh`). Every mutation run was preceded
by a GREEN no-mutation control. Ledgers: `analysis/mutation-6.1/`, `analysis/mutation-6.2/`. Scope, per Manager
ruling (B), 2026-09-28:

- (1) Mutation RED→GREEN for the covered rows that had no RED.
- (2) For every not-covered logic or parse row: a new test, or one line saying why it is unreachable, with the
  scope walked.
- (3) A committed structural test for storage-error exits instead of per-row RED. Per-row RED is
  **not-applicable**: it would need a production fault-injection seam.

The 2026-08-04 "no RED recorded" cells stay as historical fact. The RED→GREEN below is today's, recorded separately.

### Suites (`analysis/harness/run_6_1.sh`, `df` checked first; ledgers `mutation-6.1/run61-ledger-*.tsv`)

| Step | rc | s |
|---|---|---|
| untagged journal/execgw/riskbucket/officialfx | 0 | 466 |
| `tossos_testseams` execgw/riskbucket/officialfx | 0 | 67 |
| property tests + fuzz seeds (riskbucket) | 0 | 1 |
| `FuzzReservationIsMonotone` / `FuzzApplyFillRetryIsPure`, 60 s each | 0 / 0 | 62 / 61 |
| race riskbucket, officialfx | 0 | 5 |
| race execgw focus (untagged / seams) | 0 / 0 | 102 / 196 |
| race journal focus | first run hit the 10 min default timeout (no race report, no test failure); rerun with `-timeout 60m`: **0** | 984 |
| vet (untagged / seams) | 0 / 0 | — |

### (3) Storage-error exits — `TestA066StorageErrorExitsFailClosed`

This AST test asserts two properties for every storage-error exit in its scope:

- P1: the exit returns a non-nil error.
- P2: the exit performs no write (ExecContext, Exec or Commit).

It also asserts that every function which opens a transaction runs `defer tx.Rollback()` right after BeginTx.

Scope:

- every `internal/journal/risk_bucket*.go` file, plus `RecordFill`;
- `internal/riskbucket/production_snapshot_authority.go`;
- `Gateway.checkReservation` and `Gateway.submit`.

An exit counts as a storage-error exit when:

- its condition is `err != nil`, or `!errors.Is(err, sql.ErrNoRows)`;
- it sits in an if, an else-if chain (the source is that chain's init), or a `switch { … default: }`;
- and its `err` comes from a database/sql method, or from a helper that takes `tx`, `q`, `db`, `j.db` or `ctx`.

The walked scope is pinned by a census: files 10 · funcs 116 · exits 305 · others 84 (counted, not checked) ·
txOpeners 10. The census was re-counted honestly three times as the recognizer grew:

- switch default exits added: 283 → 286;
- two files added: → 301;
- not-ErrNoRows exits and else-if sources added: → 305.

Mutation: 9/9 CAUGHT (`structure-ledger.tsv`).

### 6.1.1 — owner-bind delta gap (repaired, Manager ruling 2026-09-28)

`applyRiskBucketOwnerBindingInTx` returned nil when `campaignQuantity(fill.Delta)` failed, with no owner bind and
no latch. The census surfaced it as an "other" exit.

- FLM came first (`analysis/pre-edit/6.x-owner-bind/`); RED at `47b48ae4`.
- The fix calls `latchRiskBucketFillFailureForScope`, the same latch as the sibling gaps. The fill and the exit
  hook are untouched, and a zero delta stays a no-op.
- Release semantics are unchanged.
- Mutation `owner-bind-ledger.tsv`: 4 CAUGHT. The `CompareDecimal` error branch cannot be reached; its premise is
  pinned by `FuzzA066CampaignQuantityIsComparable` (60 s, 2.5M execs, 0 failures).

### (1) Row mutation — `mutate_rows_6_1.py`, 63 rows at `a22a2c05`

Result: 52 CAUGHT, 3 CAUGHT(full), **8 SURVIVED** (`btm-rows-ledger.tsv`). Every survivor was closed
(`btm-rows-dispositions.tsv`); the re-run at the current tip is recorded below.

| Survivor | Why it survived | Closed by |
|---|---|---|
| commitFresh B15 (owner-reuse checks) | no issuance-path scale-in test | `bfaeb2bf` behavior test, 3/3 |
| Revalidate B15 (dimension loop) | three guards shadow each other (state digest / loop / `len(seen)`) | `4e5edffd` layer pin + behavior |
| fillTransition B10 (limit min) | no test with a tighter decision | `306bc872` |
| fillTransition B18 (per-decision count) | four guards shadow each other (state digest / per-decision / total / `validateFillBuckets`) | `1b46dc40` layer pin |
| ApplyFill B2 (`validateFillBuckets`) | later parse guards refused the same inputs under another reason | `53cef064`, reason field pinned |
| ApplyFill B5 (B6 loop) | B6/B32 pair; layer pin v1 accepted a head `continue` | `53cef064`, pin v2 |
| ApplyFill B25 (evidence copy) | write-only field (no production reader: `git grep` shows assignments and clone only) | `7023779d`, return contract |
| ApplyFill B36 (first-apply overflow) | `"<nil>"` stored, then refused as `overage_filled` | `e5ad6f9b`, reason field pinned |

Re-run of the 8 survivors at the tip (`btm-rows-recheck-ledger.tsv`): 6 CAUGHT at `e5ad6f9b`. Revalidate B15 and
fillTransition B18 still SURVIVED, because the row mutant injects `if true { continue }` at the loop head and both
layer pins only looked for the guard somewhere in the loop. Both pins now require the refusal to be the loop's first
statement and forbid branch statements in the loop. Recheck 2: 2/2 CAUGHT.

Pins were added only where a mutation showed shadowing (Manager rule). Two of my own pins were first too weak and
were fixed before being counted:

- the `commitFresh` B2 straight-path pin;
- the fill bucket-count pin, which v1 let `false &&` through.

### (2) Not-covered rows (158 at the 5.7 measurement)

- **81** are storage exits; the structural test above covers them.
- **Tested now:**
  - ApplyFill B4, B6, B13, B15, B17, B18, B22, B23, B24, B34, B38, and recomputeOverageLatches B5, B7, B8 —
    `47b48ae4`;
  - fillTransition B12, and Revalidate B3/B4 (B3 is a backstop whose premise `splitQFinalPolicyVersion("")` is
    pinned) — `306bc872`;
  - CRA B1, CRA B19, fresh B13 — `9ba1c783`;
  - submit **B41** (a066 `a37d97f5`) — `4773eb56`.
- **Declared backstops (behavior test plus AST pin):**
  - Revalidate B16/B17, behind `verifyRiskBucketStateDigest` — `4e5edffd`;
  - fillTransition B19, behind the state digest — `1b46dc40`.

  The 6.4 coverage refresh confirms that both still do **not execute**: the missing-dimension and
  missing-reservation behavior tests are refused by the state digest first. That is the layering the pins declare.
  They are not counted as tested.
  - fresh B2, behind `recoverQFinalIssueReplayTx` (same predicate, called first on the straight path by both
    callers);
  - ApplyFill B32, behind B6.
- **Unreachable, with the scope walked:**
  - ApplyFill B21/B27/B29/B30 and recomputeOverageLatches B2/B3/B4/B10: `validateFillBuckets`
    (`fill.go:279-307`) parses limit/held/filled/overage of every bucket and every reserved amount on entry. The
    bucket and reserved key sets are equal (equal length, containment, no duplicate dimension).
    `recomputeOverageLatches` has 2 callers (`fill.go:189`, `:272`, per `git grep internal/`), both after
    validation, and the values written in between are `big.Int.String()`.
  - fresh B4: both callers require `ExistingReservationID == soleReservationID(reserve.Reservations)`
    (`risk_bucket_issuance.go:97-98`, `strategy_first_leg_atomic.go:279-280`), and `reserveRows` inserts that row in
    the same transaction before `commitFreshRiskBucketAdmissionTx`.
  - CRA B27: the owner's reservation keys are written only by CRA/insertFresh after the identity check, and the
    state digest (a CRA scale-in precondition) comes before it. Backstop, not pinned: no mutation showed shadowing.
  - fillTransition B6: the CHECK on `bucket_dimension` (`risk_bucket.go:682`).
  - fillTransition B8: `UNIQUE(decision_id,bucket_dimension)` (`:685`), and the key includes the dimension.
  - fillTransition B27: `RegisterRiskBucketOrder` refuses the collision first
    (`TestRiskBucketOrderRegistrationRejectsBrokerOrderIDCollisionAcrossOwnerDecisions`).
  - fillTransition B32: fill rows are joined to this owner's orders, and `orderKeys` holds all of them.
  - fillTransition B13–B15, B36: stored minor strings are written only by journal code from `big.Int.String()`.
    Corruption becomes a semantic error, and the fill path latches instead of rejecting.
  - fillTransition B39: registration requires the five-dimension authority (orderAuthority B7).
  - orderAuthority B9: implied once B7 passes (`byDimension ⊆ required`, else refused at scan, and equal length).
  - Revalidate B2: `ParsePreimage` reads back the journal's own canonical preimage (written by `insertDecisionRow`).
  - JSON-digest errors — CRA B3/B32/B36, fresh B19/B21, insertFresh B3, orderAuthority B10: they marshal
    structs of string, integer, map and time fields. The times are validated fresh (year in range), so
    `json.Marshal` cannot fail.
- **Not an a066 branch** — owner commit from `git log --full-history -S'<condition text>'`; for generic
  `if err != nil` lines, from `git blame -w -M -C`, which is move/copy aware:
  - gateway.submit B2, B3, B5, B16, B27, B39, B42, B48, B52 — owner `8022f578`;
  - B30 — `9dfbfd03`; B32 — `658d4b0a`; B54 — `122985d9`; B6 — `d295555a`;
  - loadProductionRiskEntries B1, B2, B3, B6, B11, B12, B13 — `8022f578`;
  - RecordFill B10, B13 — `c93f5f4a`.

### 6.2 — exit paths (`mutate_5_6_1.py --set 6.2`, `mutation-6.2/exit-ledger.tsv`): 4/4 CAUGHT

- X01: risk-reducing decisions go through bucket authority.
- X02: a 6 s wait on the risk-reducing path (stand-in for an evidence/FX wait); the 5 s deadline bites.
- X03: reconcile entry blocked by the lock.
- X04: fill detection blocked by the lock.

The structural fact is `checkReservation`'s non-raising early return (`gateway.go:890`). X01 removes it.

### Residuals added in this lot

- **Strategy-path diagnosis flattening.** Every `call()` error on the strategy path becomes
  `ReasonStrategyDispatchFenced` (`gateway.go:715/720`, owner `8022f578`). The `entry_loss_lock_active` cause
  survives only in Detail. This is the same family as the worker-wide latch residual, and joins the follow-up batch
  of the activation-wiring lot (Manager 2026-09-28).
- **Bundle correction.** `readProductionRiskUsage` and `aggregateProductionRiskUsage` were **not** deleted in 5.6.1
  (they are at `production_snapshot_authority.go:450/473` with new signatures). The earlier "retire" was withdrawn;
  they are refresh targets in 6.4.

### 6.4 coverage refresh (2026-09-28, package-wide at `97ea352e`)

- Measurement: journal untagged `-coverpkg journal,riskbucket` (rc 0, 460 s, 72.1 %); riskbucket and execgw with
  `tossos_testseams` (rc 0).
- The measured rows appear in the BTMs **alongside** the historical cells (Manager 2026-09-28: add alongside, no
  replacement), in 15 bundles.
- Rows that were NOT covered: 188 in total, including the 6.1.1 bundle's focus-set rows. **26** are now covered;
  **162** are still not covered.
- The 162 split as follows: storage exits (structural test); declared backstops (pinned); unreachable rows with a
  walked-scope line; branches not owned by a066; and the 6.1.1 bundle rows outside the focus set. Each has its line
  in the (B)(2) list above.

### Gate probe (2026-09-28, Manager slot) — isolated worktree `TossOS-worktrees/a066-gate` @ `63c51e00`

- Setup: `make sdd-infra` rc 0 (the venv is local to the worktree).
- `make sdd-sync` rc 2. The CodeGraph hard-evidence sync completed. The CodeGraphContext advisory could not take
  its database lock (`~/.codegraphcontext/global/db/kuzudb` is held by a running `cgc mcp start`, not this lot's
  process, and was left alone).
- `make sdd-check` rc 0: "CodeGraph hard-evidence index matches the worktree"; the advisory indexes WARN.
- **`make gate CHANGE=a066-…` rc 2, stopped at step 2/11** with 4 open tasks, as expected:
  - 5.5: relaxation of the loss lock and of the overage latch waits on the user;
  - 6.3, 6.4, 6.5: this lot. 6.5 runs after this probe, on the final tree (Manager ruling).
- Lot verification bundle in the same worktree:

| Command | rc | s |
|---|---|---|
| `make lint` | 0 | 23 |
| `make test-seams` | 0 | 680 |
| `make sdd-test` | 0 | 139 |
| `openspec validate --strict` | 0 | — |
| `check_analysis --change a066` | 1 | 19 |

- The `check_analysis` rc 1 is 371 "missing evidence" lines. None of them is a function changed by an a066
  production commit since the base (`0004536c 7aa158bb b8211926 bb44e7af 2b36ae44 9f9aa8e3`). The only overlap by
  file is `gateway.go`, and `0004536c`'s +5 lines there are in `checkReservation`, which has a bundle. The rest come
  from other changes' commits in the stacked window from base to worktree (the same artifact recorded in Wave 2A).
- `make test` rc 0 (658 s) and `make vet` / `make validate` rc 0 were measured on the same commit in the main
  worktree (6.4 above).

## 6.5 independent adversarial review (2026-09-28, final tree `f1094a62`)

Four voices ran read-only against the final tree. Probes ran in throwaway copies, which are now deleted, and each
voice checked that the repo state was unchanged afterwards.

- **R1** — Claude `code-reviewer`: atomicity, monotonicity, owner races.
- **R2** — Claude `code-security-auditor`: exit bypass and the lock.
- **R3** — Claude `general-purpose`: the old-wave removals and whether anything loosens a limit.
- **CX** — Codex (`gpt-6-astra`, challenge mode, read-only sandbox, about 2.4M tokens).

No voice found a way for a066 to refuse, delay, or abort a stop, an emergency exit, a reduce-only cancel, a
reconcile entry, or a SELL fill on a healthy journal. Probes confirmed this. The exception is the corruption-only
fill abort in row 4. R3 found no a066 commit that loosens an existing limit.

| # | Finding | Voices | Severity (max) | Proposed disposition |
|---|---|---|---|---|
| 1 | `RevalidateQFinalAdmission` does not re-check latches: owner, scope, and shared-bucket `Latched`. An already-issued q_final entry is submitted after its owner (resealed digest) or a shared bucket latched, while a fresh admission is refused. Probe-confirmed by R1 and R2. | CX1 · R1-2 · R2-1 | P1 | **Fix in this lot (proposed)**: add a latch check as the last step of revalidation, next to the lock check (the same "exposure is the state at submit time" rule as decision ⑤). Entry path only. |
| 2 | `filled_minor` is never released. Every bucket becomes a lifetime cumulative cap, and released owners still count. The owner-lifecycle fixture hides this with `filled_minor='0'`. Probe-confirmed by R1. | CX7 · R1-1 | P1 (fails closed) | **Stop and report**: this is release semantics (design D5, "filled exposure attributed to the Position projection"). A user or Manager design decision. |
| 3 | A shared bucket's limit is per-admission (each entering snapshot's own limit). Two manifests can declare different sector or strategy limits, and horizon limits come from separate KR and US manifests. The effective cap is then the largest one. Probe-confirmed by R1. | CX5 · R1-3 | P1/P2 | **Decision needed**: (a) at admission, cap by the smallest limit recorded on the bucket's existing reservations; or (b) upstream manifest validation fixes one limit per shared bucket value. The same family as the upstream policy-immutability residual. |
| 4 | Missing state-snapshot seal: `verifyRiskBucketStateDigest` returns a raw `sql.ErrNoRows`, which is not classified as semantic, so the whole fill transaction aborts. R2 found a few more corruption-only raw returns (`riskBucketFillEventDigest`, `releaseRiskBucketOrderInTx` held read). | CX6 · R2-4 | P1 (CX) / P3 (R2) | **Fix in this lot (proposed)**: classify these reconstruction gaps as semantic, so they latch and the fill commits (the 2.7 contract: fill detection is never blocked). |
| 5 | Owners that never filled, and strategy-dispatch releases, which break the state digest because they don't reseal: the owner can never be released, and the symbol stays locked. | CX4 · R1-4 | P1/P2 (latent) | Residual → activation-wiring lot. Owner release is unreachable in production today. |
| 6 | Production `PolicyVersion` is unique per collection, so same-owner scale-in always fails "scale-in bucket identity". The fresh-scale-in control uses a fixed version. | CX3 · R1-6 | P1/P3 (fails closed) | Residual → activation lot. The control test gets a note: its fixed version is not a production shape. |
| 7 | Production snapshot reader pinned to schema 27. | CX2 | — | Already a named residual (5.6.1). |
| 8 | A late fill on a released owner does not latch the shared buckets. | R1-5 | P2 (latent) | Residual → activation lot (owner release). |
| 9 | Reconcile: `EnterReconcile` for a global row fails (v24 trigger) when a market-scoped row is active, and `Tracker.persist` stops at the first error. The legacy writer's dedup ignores `scope_market`. | R2-2 · R3-1 | P2/P3 (latent) | Residual → activation lot. The only production writer of market-scoped rows needs owner release. |
| 10 | The non-strategy q_final path has no production order registration. | R1-7 | P3 (latent) | Residual → the lot that wires it. |
| 11 | The non-owned-fill latch matches order id without side or day (a conservative false positive). | R2-3 | P3 | Record only. |
| 12 | The strategy path flattens the loss-lock reason (Detail keeps it; nonce spent; no retry to send). | R2-5 | P3 | Already a residual (activation-lot follow-up). |
| 13 | `official.New` uses a private transport, identical to `http.DefaultTransport`. That escapes `testenv.InstallGuard`, which has 0 callers. | R3-2 | P3 (latent) | Record only (not a066 code risk). |
| 14 | Stale comment `symbolgate.go:180`; market-scoped rows cannot be released by production callers (they over-block). | R3-3 | P3 | Record only. |

Outside a066: Amend's `raisesExposure` ignores side (`gateway.go:447`, R2) — recorded for its owner. `8022f578`
removed the a090 mixed-currency refusal, replaced by account-base FX, and does not loosen anything (R3).

### 6.5 dispositions (Manager ruling 2026-09-28) and repairs

- **#1 submit revalidation re-checks latches: repaired** (`28629ec6`).
  - `RevalidateQFinalAdmission` ends with the same rule functions admission uses:
    - `ensureRiskBucketEntryScopeClean`: owner latch, scope latch, active reconcile. It now names the cause.
    - `latchedUsageRefusal` for each of the decision's buckets, naming the bucket and the latch kind.
  - The reason code `guardian_risk_bucket_mismatch` is reused. The cause is in Detail, as the Manager ruled: unlike
    ENTRY_LOSS_LOCK, the substance (RISK_OVERAGE and the like) already has its own name in the ledger's latch rows and
    can be looked up there, so naming it in Detail lets the refusal say its own name without a new Gateway code.
  - Gateway test `TestA066IssuedEntryIsRefusedAtSubmitWhenItsScopeLatchedAfterIssuance`: broker 0, NOT_DISPATCHED.
- **#3(a) shared bucket limit: repaired** (`28629ec6`, then `cb16367d`).
  - At admission, the ledger comparison site caps by the smallest snapshot limit recorded on the bucket's rows. The
    rows are the ones the ledger counts: HELD, FILLED, and RELEASED with filled usage.
  - **#3(b)**, one limit per shared value in upstream manifest validation, is a named residual in the same family as
    the policy-immutability residual.
- **#4 missing state seal: repaired** (`28629ec6`). `verifyRiskBucketStateDigest` types it as a replay mismatch, so
  the fill path latches and the fill commits.
  - The other two raw returns R2 listed sit behind this check on every path, so they are backstops and were not
    edited.
- **#2 usage lifecycle: stopped and reported** (user queue). The owner-lifecycle fixture now carries a comment saying
  that `filled_minor='0'` hides the gap.
  - The narrowed re-review adds a related liveness residual: because FILLED rows keep their recorded limit forever,
    raising a limit never takes effect on a bucket that has history. This is part of the same user decision.
- **#5–#14**: named residuals, as in the table above. For #6, a comment on the fresh-scale-in control says its fixed
  policy version is not a production shape.
- **Verification.**
  - FLM first (`analysis/pre-edit/6.5-fixes/`, `6.5-fixes-2/`).
  - RED at `4fb0b26e` (journal 3, Gateway 1) and at `28629ec6` (2).
  - Mutation (`analysis/mutation-6.5/`):
    - fixes: 18/18 CAUGHT, after two first-pass survivors — V09 (owner overage) and L03 (smallest vs largest) —
      each got a test;
    - re-review: 3/3 CAUGHT.
  - journal/execgw/riskbucket untagged rc 0, seams (execgw, riskbucket, engine) rc 0, vet rc 0.
  - Census updates: storage exits 305 → 314; the ledger call census now includes the revalidation site.
- **Narrowed re-review of `28629ec6`** (one voice, `code-reviewer`). It found P2 (RELEASED-with-filled rows missing
  from the limit population), P3 (a reconcile row with an empty cause did not block) and P3 (no test for the
  owner/scope-latch branches at the submit site). All three were repaired or tested in `cb16367d`. P3 liveness goes to
  the usage-lifecycle residual. It confirmed there is no exit-path reach, no bypass, and no loosening from the
  typed-seal change.

## 5.5 relaxation lot (2026-09-28/29) — mechanism, review and repairs

User decision 2026-09-28 (relayed by Manager): automatic paths only tighten. Relaxation needs OPERATOR, an approval
reference, and an audit line before commit. The entry points are the journal API and tossctl `mutating: true` commands;
there is no console button. This lot builds the mechanism only: nothing runs on the operating journal and nothing is
activated. Design D8.

### Commits

- `bce793a7` — v35 (REAFFIRM events, lock releases, `one_open` trigger swap, latch releases); journal
  `ReleaseEntryLossLock` / `ReleaseRiskOverageLatch`; tossctl `entry-lock-release`, `risk-latch-release` (mutating) and
  `risk-latch-show`; frozen audit-action census.
- `75d9073b` — analysis debt that the 6.5 repairs left behind: 11 stale ASTs, FLM rows B20–B27, missing sections.
- `90e5170d` — the a092 command-family contract check found that the CLI wrote the single-writer journal through
  `journal.Open`, which migrates. The releases now go through the engine control endpoint, discovered as an a079-style
  optional capability. The engine writes with its own journal and audit log. The CLI refuses when the engine is not
  running (Manager Q2(a)). A post-commit `engine.risk_relaxation` notice goes out, and its failure is reported as
  「완화됨·통지 실패」.
- `48b00df5` — CX-1: a typed-nil `(*audit.Log)(nil)` Auditor was accepted and committed without an audit line (RED,
  then repaired).
- The repair commit after this section — review repairs listed below.

### Independent review at `90e5170d` (raw reports: `analysis/review-5.5/`)

Four voices, all read-only, probes in deleted copies: R1 safety (`code-security-auditor`), R2 endpoint/CLI
(`code-reviewer`), R3 evidence (`general-purpose`), and CX (Codex, read-only sandbox).

No voice found a path that relaxes without OPERATOR + approval + audit before commit. None found anything that delays or
refuses a stop, an exit, reconciliation, or fill detection. None found a release that clears another owner's latch; R1
probed the shared bucket. The single writer, auth, and descriptor ordering hold (R2).

| # | Finding | Voices | Sev | Disposition |
|---|---|---|---|---|
| 1 | Typed-nil auditor accepted by the journal API → unaudited commit | CX · R1 · R2 | P0 (CX) / P2 | **repaired** `48b00df5` (`nilAuditor`); M08/M23 CAUGHT |
| 2 | Audit fsync inside the transaction holds the single journal connection | CX · R2 · R1 · R3 | P0 (CX) / P2–P3 | **Manager Q6 (i)**: kept. Measured 4.6 ms audit / 9.1 ms release (p95 10.7 ms); named residual in D8 |
| 3 | Commit failure after the audit line → audit says "released", journal rolled back | R2 (probe 2) · R1 · R3 | P3 | **repaired** (Q6 (i)): the pre-commit value is `release_attempt`, and a `not_committed` line is written when the commit fails (independent of the request context). Test `TestA066AuditLineBeforeCommitIsAnAttemptAndACommitFailureIsCompensated`; M27/M28 CAUGHT |
| 4 | Account number in the externally published notice | R2 | P2 | **repaired**: Title/Body carry market/scope only. Same §8 family as a090 r2 (user queue) |
| 5 | Refusals that change nothing were reported as "outcome unknown" (seal mismatch; audit write failure) | R1 · R2 · R3 | P2/P3 | **repaired**: named `state_mismatch` / `audit_unavailable`; E11/E12/CX19 CAUGHT |
| 6 | `risk-latch-show` on a pre-v35 journal → raw SQL error | R2 | P3 | **repaired**: `ErrSchemaTooOld` (test `TestReadOnlyRelaxationViewsRefuseAPreV35Journal`) |
| 7 | Evidence debt: v35 immutability and UNIQUE, route auth, latch-path nil audit, generation filter, overage_minor kept, released/unknown owner, show read-only, ReadOnly values, census bypasses (local shadow, method value, local const), other-scope binding, M19 behavioural | R3 | P2 | **tested**: `a066_relaxation_evidence_test.go`, engine auth / latch audit tests, CLI show test, hardened census. MX01–MX07, MX12, MX17, MX18, MX20, EX13, EX14, EX16, CX21, M24–M26 all CAUGHT |
| 8 | Harness counted a build failure as CAUGHT (M22) | R3 | P2 | **repaired**: `BUILD-FAILED` verdict; M22 now compiles and is CAUGHT |
| 9 | Stale `status.md` / `tasks.md` sentences | R3 | P2 | **repaired** |
| 10 | Notice is a direct recorder outside a092's recording entry | a092 census · Manager | — | **Q5 (a)**, binding after a092 r23: migrate onto `RecordAlert` when a092 lands it (tasks §7 named hand-over; enforced by a092 tasks 24.4). Mutual precondition with a092's archive |
| 11 | REAFFIRM on every activation could make approvals always stale if a trigger fires per cycle | R1 | P3 | residual → trigger lot (cadence and deduplication) |
| 12 | After releasing one owner's latch on an over-limit shared bucket, entries are refused by the cap, not by the latch | R1 | P3 | record (the safety outcome is unchanged) |
| 13 | `s.mu` held across the transaction and the notice | R2 | P3 | record (the quarantine release needs the same connection anyway) |
| 14 | The engine does not log a notice failure | R2 | P3 | record |
| 15 | "Scope latch rows untouched" is declared only | R3 | P3 | record (the release SQL touches only owners/reservations) |

### Mutation (`analysis/harness/mutate_5_5_relaxation.py`, copies with pid, a no-mutation control per suite)

- `journal-ledger.tsv`: the first run, kept as history. M11/M20/M21 survived and got tests.
- `ledger.tsv`, `rerun-ledger.tsv`: after the socket switch. E10/C04 survived and got tests.
- `ledger-r3.tsv` + `ledger-r3-audit-sites.tsv`: after the review repairs. **61/61 CAUGHT** across journal M01–M28 and
  MX, engine E01–E12 and EX, and CLI C01–C06 and CX. The three control suites are GREEN, with no build failures. M09,
  M10 and M16 were re-anchored to the attempt line and rerun.

### Narrowed re-review of `93f74c79` (one voice, `code-reviewer`, 2026-09-29)

Verdict: the runtime repairs are correct and complete. No path says "nothing was released" for a release that committed.
Probe: 2000 releases with a random cancel gave 0 violations of err==nil ⇔ released, and 0 `not_committed` lines on a
committed release. The state_mismatch and audit_unavailable errors are produced strictly before `Commit`.

- **P2 (introduced by the census hardening): repaired.** The census walked only `FuncDecl` bodies and so missed
  `RecordAction` inside package-level `var` initializers. It now walks the whole file for calls and method values or
  expressions, and keeps function-body scope only for local shadows. M29 (package var func literal) and M30 (package
  method expression) are CAUGHT, and M24–M26 are re-CAUGHT (`ledger-rereview-census.tsv`).
- P3 records:
  - `nilAuditor` sees one level only; a value wrapper around a nil `*audit.Log` passes. Latent: production passes
    `*audit.Log` directly.
  - An attempt line can be left without a compensation line when the audit fsync fails after the write, on a process
    crash, or when the compensation write itself fails. The ledger is unchanged and the CLI's "refused" is true.
  - After a WAL fsync error at commit, a `not_committed` line could sit beside a durable commit. The CLI then says
    "outcome unknown", which is the right direction.
