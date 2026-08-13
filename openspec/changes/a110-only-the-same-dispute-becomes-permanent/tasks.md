# Tasks — a110-only-the-same-dispute-becomes-permanent

## 0. Manager evidence and contract freeze

- [x] 0.1 Read `.claude/CLAUDE.md`, `docs/WORKFLOW.md`, canonical reconciliation/exit/operator specs and prior a062/a083 decisions.
- [x] 0.2 Reproduce the live symptom read-only: desired/effective adoption ON, three candidates covered by one account-wide permanent block, no baseline/snapshot for those candidates.
- [x] 0.3 Correlate the 2026-08-07 journal and engine log: three changing symbol disputes pooled into one scalar streak; two ordinary blocks later released with `ADJUSTMENT_APPLIED` while the account-wide permanent row remained.
- [x] 0.4 Run `make sdd-sync`; record CodeGraph definition/callers/callees/impact for `Tracker.Observe` and the adoption gate.
- [x] 0.5 Generate pre-proposal AST, Function Logic Map, Branch Test Map and risk report for `Tracker.Observe` and supporting `ReconcileDriver.blocked`.
- [x] 0.6 Reserve `STORY-TOS-a110`, capture frozen base, write proposal/design/delta spec/tasks.
- [x] 0.7 Run proposal-freeze gstack review with an adversarial Eng voice, resolve every P0/P1, record decisions in `review.md`, and strict-validate before implementation.

## 1. T1 RED — promotion identity unit contract

- [ ] 1.1 Add a dedicated a110 test file owned by T1 proving three changing quantity symbols do not pool into permanent promotion while every ordinary block latches immediately.
- [ ] 1.2 Prove the same symbol with a changed canonical local/broker tuple starts a new streak, while equivalent decimal spellings continue the same streak; include a float64 2^53 collision pair that must remain distinct.
- [ ] 1.3 Prove one exact quantity dispute observed at threshold still creates the existing account-wide, durable, operator-only permanent block.
- [ ] 1.4 Prove duplicate copies of one identity inside a comparison count once and clean comparison clears every transient streak; blank, malformed, NaN/Inf items remain ordinarily blocked but earn zero streak, while a valid sibling key still counts.
- [ ] 1.5 Prove permanent enter failure then clean/different key withdraws the non-durable account pending retry, while the same key retries and becomes durable; assert pre/post persistence gate ordering and ordinary pending retry preservation.
- [ ] 1.6 Run the focused tests against the frozen implementation and record the intended RED failures without changing production code.

## 2. T2 RED — missing-order and incident-chain contract

- [ ] 2.1 Add canonical missing-order streak tests: identical full identity reaches threshold; reused opaque ID across market/day/symbol/side cannot share a streak; each of the six blank components keeps ordinary blocking but earns zero streak/no permanent, while a valid sibling still counts.
- [ ] 2.2 Add the sanitized 2026-08-07 sequence: different symbol mismatches arrive over three cycles, earlier blocks earn adjustment credits and release, and no account-wide permanent block survives.
- [ ] 2.3 Add an engine-level regression proving the changing-dispute sequence does not suppress quote/adoption after ordinary blocks clear; adoption first persists t0 and a `SEED` exit state, the intermediate view stays `not_evaluated_yet`, and a separate exit-observer judgement persists the evaluated snapshot before lines become actionable. No LIVE order adapter may be wired.
- [ ] 2.4 Preserve the existing permanent-gate regression: an already durable permanent block still stops every adoption candidate before a price read.
- [ ] 2.5 Run the owned focused tests against the frozen implementation and record the intended RED failures without changing production code.
- [ ] 2.6 Prove two transient observations followed by Restore or Refresh do not manufacture/reconstruct a third streak; an already durable permanent row still restores as permanent.

## 3. T3 GREEN — minimal streak implementation

- [ ] 3.1 Add an unexported comparable promotion key that uses exact finite-decimal canonicalization without float64 for quantities and reuses `LocalOrder.Identity()` normalization for orders while requiring all six canonical order components non-empty.
- [ ] 3.2 Replace identity-free promotion accounting with a bounded current-dispute streak map; derive `failures` as the maximum current streak for compatibility.
- [ ] 3.3 Count each identity once per comparison, drop absent identities, reset all transient streaks on non-blocking comparison, and keep invalid decimal or incomplete-order items ordinarily blocked without granting permanent-streak evidence.
- [ ] 3.4 Promote through the unchanged durable account-wide operator-only path when any one exact streak reaches the existing threshold.
- [ ] 3.5 Keep B8–B21 ordering intact: pre-persist gate latch, exact-cause journal writes/releases, authoritative conflict replacement and credit accounting.
- [ ] 3.6 Preserve restart/refresh behavior: durable permanent rows survive; transient streaks are not reconstructed; no clean observation auto-releases permanent state.
- [ ] 3.7 Bind a failed account-wide permanent write retry to its earning key's immediately consecutive blocking observation; withdraw only stale non-durable account pending promotion on clean/key disappearance, never ordinary pending or durable permanent blocks.

## 4. Teammate verification and mutation ledger

- [ ] 4.1 T1 and T2 rerun all owned tests GREEN against T3, then run `go test ./internal/reconcile ./internal/app/engine` with appropriate timeout.
- [ ] 4.2 Run focused `-race` tests for reconcile and the a110 engine slice; run `go vet` on affected packages.
- [ ] 4.3 Mutation: restore the scalar shared counter and prove the incident test RED; change quantity key to symbol-only and prove tuple test RED; drop one missing-order scope field or its non-empty validation at a time and prove canonical/incomplete-identity tests RED.
- [ ] 4.4 Mutation: allow clean comparison to remove permanent state and prove existing permanent/operator tests RED; bypass ordinary block while identity is unclassifiable and prove fail-closed test RED.
- [ ] 4.5 Regenerate post-edit AST/maps/risk report and make `check_analysis.py --change a110-only-the-same-dispute-becomes-permanent` pass.

## 5. Independent adversarial review

- [ ] 5.1 A1 (different Terra context) reviews promotion identity, canonicalization, counter reset, duplicate handling and restart semantics; records P0/P1/P2 with concrete counterexamples.
- [ ] 5.2 A2 (different Terra context) reviews durable ordering, gate scope, adoption suppression, operator-only release and exit immediacy; must try to make a false clean or false permanent state.
- [ ] 5.3 T3 fixes every accepted P0/P1 with new RED-first coverage; original reviewers re-review their findings to closure.

## 6. gstack and Manager completion gate

- [ ] 6.1 Run gstack code/Eng/security/QA review after adversarial closure and record decisions and any requirement edits in `review.md`.
- [ ] 6.2 Run `openspec validate a110-only-the-same-dispute-becomes-permanent --strict --no-interactive`, `make sdd-sync`, `make sdd-check`, affected tests/race/vet, then `make gate CHANGE=a110-only-the-same-dispute-becomes-permanent`.
- [ ] 6.3 Manager independently inspects the final diff, tasks↔design↔delta alignment, mutation evidence and frozen-base/post-edit AST changes; completion requires P1=0.
- [ ] 6.4 Record the operational boundary: merge/deploy and the current block's audited release remain separate human approvals; no order or operating toggle was invoked by this change.
