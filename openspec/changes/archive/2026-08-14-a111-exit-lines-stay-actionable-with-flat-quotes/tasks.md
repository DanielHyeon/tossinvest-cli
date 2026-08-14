# Tasks — a111-exit-lines-stay-actionable-with-flat-quotes

## 0. Manager evidence and contract freeze

- [x] 0.1 Preserve the read-only A110 production receipt: exact image digest, healthy services, schema 31, active reconcile 0, five adoptions, three `SEED` and the evaluated rows that aged stale under flat quotes.
- [x] 0.2 Run memory recall and CodeGraph definition/callers/callees/impact for `ExitObserver.judgeRatchet`, `judgeLadder`, `record`, journal exit judgement persistence and shared operator freshness projection; record hard evidence before implementation.
- [x] 0.3 Generate pre-edit Go AST, Function Logic Map and Branch Test Map for both judge functions, the new/affected journal transaction and console/API freshness consumers; no high-risk function is exempt.
- [x] 0.4 Reserve `STORY-TOS-a111`, record the frozen main/image/schema baseline, and align proposal, design, three delta specs and this task ledger.
- [x] 0.5 Have a separate Terra adversary review the frozen proposal/design/spec/tasks for event semantics, transaction concurrency, freshness authority and order safety; resolve every P0/P1 and record the final P1=0 verdict in `review.md`.
- [x] 0.6 Run `openspec validate a111-exit-lines-stay-actionable-with-flat-quotes --strict --no-interactive` before any RED or production edit.

## 1. T1 RED — observer lifecycle and side-effect contract

- [x] 1.1 In a dedicated T1-owned engine test file, prove a ratchet `SEED` receiving its unchanged first valid quote becomes `EVALUATED` with canonical current/next lines, one semantic event and zero proposal/cancel/issuer/submitter calls.
- [x] 1.2 Prove the same unchanged-first-quote behavior for an automatically adopted common-ladder position using the real adoption → exit-state → separate ExitObserver chain and the same no-side-effect counters.
- [x] 1.3 Prove a second and repeated identical fresh quote advances durable observation ID/observed-at and remains fresh beyond 30 seconds without adding an exit event, proposal, intent, cancel or order.
- [x] 1.4 Prove an evaluator-generated non-scalar operational difference missed by legacy high-water/protection/level detection cannot use refresh: pending suppression changes action/ratio/projection/orderability through the complete judgement path while retaining proposal dedupe and recovery rules; separately preserve that a quantity change which alters projection remains a full judgement.
- [x] 1.5 Prove a real protection breach and pending/re-judgement cases retain durable-record-before-submit, single-proposal dedupe, protective immediacy and existing recovery selection.
- [x] 1.6 Prove missing, malformed, non-finite, future and source-stale quotes do not advance persisted observation evidence or snapshot freshness.
- [x] 1.7 Prove non-zero fetched-at uses one post-batch engine clock: future is rejected, exactly 15 seconds old is accepted, older is rejected; zero fetched-at uses only the explicit cycle fallback and a bad non-zero stamp cannot fall back.
- [x] 1.8 Prove a restarted observer recovers the maximum persisted `cycle:N` from the working set before allocating the next fallback, including a frozen/coarse clock fixture.
- [x] 1.9 Prove an HTTP-success all-invalid batch is a typed non-retryable evidence failure: no second broker call, no Retrier `QueryPrice` success/freshness clear, no outage reset and no exit mutation; valid siblings still allow only valid-symbol judgement.
- [x] 1.10 Use a slow first-position seam and multiple managed positions to prove the later candidate's quote is rechecked before judge and immediately before record/refresh/clear; expiry after the first read causes zero later side effects and no extra price call.
- [x] 1.11 Prove every cycle retains one batched Prices call for the complete managed symbol set, with no extra broker read and unchanged fill-detection priority.
- [x] 1.12 Run only the owned focused tests against frozen A110 and record the intended RED failures; preservation controls for real changed-price judgement and order safety must remain GREEN.
- [x] 1.13 Prove a wall-clock rollback cannot extend the 15-second use lease: monotonic elapsed time over the bound, or wall time moving before fetched-at, skips the later position before judgement/record/refresh/clear/order side effects.

## 2. A1 adversarial review of T1

- [x] 2.1 A separate Terra reviewer inspects T1 without editing it and attempts to pass with a SEED-only special case, hidden event spam, fake clock refresh, extra quote read, or accidental submit/cancel side effect.
- [x] 2.2 T1 fixes every accepted P0/P1 in its owned test file RED-first; A1 re-runs focused tests and records P0=0/P1=0 before production implementation.

## 3. T2 RED — journal refresh and shared projection contract

- [x] 3.1 In T2-owned journal tests, specify a narrow refresh transaction that accepts only an `EVALUATED`, same-lifecycle, noncompleted, nonorderable candidate whose complete operational line equals the current effective snapshot.
- [x] 3.2 Prove refresh atomically replaces JSON plus every flattened snapshot/provenance/observation field, is idempotent for one observation, survives restart and appends no `exit_events` row.
- [x] 3.3 Prove refresh rejects `SEED`, proposal/arm evidence, generation/policy mismatch, semantic line difference and a concurrent stronger high-water/protection/rung without changing any state column.
- [x] 3.4 Inject transaction failures at the state-write boundary and prove no partial tuple, event, proposal or false freshness is committed.
- [x] 3.5 Prove reverse commit order cannot replace a newer observation with an older tuple; exact same time+identity delivery is a true no-op including unchanged `updated_at`, while same time+different identity is rejected.
- [x] 3.6 Prove equal-time zero-fetched-at refreshes use persisted `cycle:<sequence>` ordering under a frozen clock; lower sequence, same sequence+different identity and official same-time/different identity are rejected, while same sequence+identity is a no-op.
- [x] 3.7 Race/replay two journal refreshes that issue the same `cycle:N` and timestamp with different in-band prices/identities; exactly one tuple wins and the loser changes no state, `updated_at` or event.
- [x] 3.8 Prove equal-time source precedence: official supersedes cycle, cycle cannot supersede official, and distinct official identities conflict; restart-resumed cycle ordering remains durable.
- [x] 3.9 Add shared operatorview/console/httpapi projection tests: running/unavailable/unwired is fresh at 29.999 seconds and exactly 30 seconds, stale only over 30 seconds; stopped is immediately `engine_not_running` on both surfaces. On actual routes, cross the same boundaries during cache/journal/runtime/quarantine/marker reads and require one post-marker response clock; a pre-read stopped marker cannot be resurrected by clock rollback.
- [x] 3.10 Prove a running engine with one repeatedly invalid/missing symbol cannot borrow a valid sibling's liveness: only that symbol ages stale after 30 seconds on console and API.
- [x] 3.11 Prove corrupt/future/partial tuples and runtime-running without a valid persisted observation remain hidden; `SEED` raw evidence stays non-actionable.
- [x] 3.12 Run only T2-owned focused tests against frozen A110 and record intended RED failures while existing snapshot integrity/quarantine/event tests remain GREEN.

## 4. A2 adversarial review of T2

- [x] 4.1 A separate Terra reviewer inspects T2 without editing it and attempts timestamp-only writes, JSON/flattened divergence, stale overwrite races, saved-monotone weakening, event duplication and console/API clock divergence.
- [x] 4.2 T2 closes every accepted P0/P1 RED-first; A2 re-runs focused tests and records P0=0/P1=0 before production implementation.

## 5. T3 GREEN — minimal production implementation

- [x] 5.1 Add the shared observer outcome classifier so `SEED`/missing-effective and every semantic change use the existing full judgement path, while only an exact evaluated flat result can request refresh in both ratchet and ladder.
- [x] 5.2 Add the narrow journal refresh request/transaction with existing snapshot validation, lifecycle/generation/status checks, exact operational semantic equality, atomic complete-tuple replacement and typed conflict/no-write behavior.
- [x] 5.3 Add temporal CAS: official evidence rejects older or same-time-different observations; exact duplicates are no-op without `updated_at` churn; equal-time cycle fallback uses persisted monotone `cycle:<sequence>` ordering with official-over-cycle precedence.
- [x] 5.4 Recover the observer sequence from the maximum valid persisted `cycle:N` in the current working set before allocating a fallback, including after restart.
- [x] 5.5 Keep refresh free of event append, proposal arm, clear/cancel, issuer, submitter, outbox and LIVE adapters; keep first flat `SEED` evaluation as one normal nonorderable judgement event.
- [x] 5.6 Validate quote source time once after the existing batch read with exact 15-second/future/zero-fallback rules before any judgement, without adding a broker call.
- [x] 5.7 Make all-invalid evidence a typed non-retryable Retrier failure before success bookkeeping; keep UTC fetched-at provenance separate from a process-local monotonic lease anchor, and recheck both wall provenance and exact 15-second elapsed time before each judge and record/refresh/first side effect.
- [x] 5.8 Make the persisted effective observation time mean latest successful line revalidation and align integrity comments/readers without changing schema 31.
- [x] 5.9 Route console and HTTP positions through one operatorview freshness decision; supersede a077's running age bypass now that the timestamp is a heartbeat, apply the exact 30-second bound to running/unavailable/unwired and retain stopped-immediate-stale. Capture response authority after all blocking reads and the single marker read, and never upgrade a marker already classified stopped.
- [x] 5.10 Preserve one quote batch, rate priority, monotone selector, pending/re-judgement paths, protective immediacy and durable-record-before-submit ordering; once an orderable decision begins valid, do not abandon a half-cleared/armed path.
- [x] 5.11 Run T1/T2 focused suites GREEN, affected package suites and existing adoption/snapshot/quarantine/order regressions; production edits remain limited to the mapped functions and supporting journal/view seams.

## 6. A3 adversarial review of T3

- [x] 6.1 A third Terra reviewer independently inspects the production diff and transaction ordering, traces every new branch to a RED, and supplies executable counterexamples for any P0/P1 finding.
- [x] 6.2 T3 fixes accepted P0/P1 only after T1/T2 add the corresponding RED; A3 re-reviews to P0=0/P1=0.

## 7. Mutation, performance and teammate verification

- [x] 7.1 Kill mutations that remove the SEED full-record exception, restore the flat early return, turn refresh into full judgement/event append, or permit orderable refresh.
- [x] 7.2 Kill mutations that compare only state scalars, update timestamp without JSON/IDs, omit one flattened field, ignore lifecycle generation, accept same-cycle-sequence/different-identity replay, or overwrite concurrent stronger state.
- [x] 7.3 Kill mutations that let invalid/stale quotes refresh, mark an all-invalid batch successful, replace monotonic elapsed time with rollback-prone wall deadlines, omit the per-position use-time recheck, let running runtime bypass age/integrity, resurrect a pre-read stopped marker after clock rollback, let a valid sibling refresh an invalid symbol, break cycle-sequence ordering, or give console and HTTP different freshness clocks/bounds.
- [x] 7.4 Run focused tests with repetition and `-race`, full affected package tests, `go vet`, and bounded flat-loop/write-count checks proving no extra broker read and no event growth.
- [x] 7.5 Regenerate post-edit AST/Function Logic Map/Branch Test Map/risk report and make `check_analysis.py --change a111-exit-lines-stay-actionable-with-flat-quotes` pass.

## 8. gstack and Manager completion gate

- [x] 8.1 Run gstack code, Eng, security, data-integrity, QA and red-team review after A1/A2/A3 closure; record all findings and dispositions in `review.md`, with completion requiring P0=0/P1=0.
- [x] 8.2 Run `openspec validate ... --strict`, `make sdd-sync`, `make sdd-check`, full tests/race/vet and the final `make gate CHANGE=a111-exit-lines-stay-actionable-with-flat-quotes` command; any gate failure reopens this task.
- [x] 8.3 Manager independently verifies final diff, task/design/spec alignment, journal atomicity, event/proposal/order counts, mutation ledger, regenerated maps and console/API parity; Manager does not implement code or tests.
- [x] 8.4 Update PM/story evidence and record archive/sync, merge, exact-digest deployment and any operational toggle as post-gate human approval boundaries.

## 9. Post-gate human approval boundary (not an implementation task)

The following operations are deliberately outside this change's completion checklist. After a new explicit
human merge/deployment approval, archive/sync the accepted OpenSpec, capture a dormant exact-digest preimage
and schema/rollback compatibility, then replace `httpapi` before `tossos` one service at a time without
`down/up` or mutable tags. Verify exact running digest, health, schema 31 and active reconcile 0 after each
replacement; rollback only the applied subset on a failed bound. On the first valid engine cycle, prove every
managed `SEED` row becomes `EVALUATED`, then observe a flat position for more than 30 seconds and prove journal,
console and HTTP lines remain fresh with no flat-only exit event/proposal/order growth.
