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
