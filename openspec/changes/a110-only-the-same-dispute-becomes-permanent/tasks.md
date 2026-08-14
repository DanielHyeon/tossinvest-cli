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

- [x] 1.1 Add a dedicated a110 test file owned by T1 proving three changing quantity symbols do not pool into permanent promotion while every ordinary block latches immediately.
- [x] 1.2 Prove the same symbol with a changed canonical local/broker tuple starts a new streak, while equivalent decimal spellings including canonical zero continue the same streak; include a float64 2^53 collision pair through the real `Comparer` path that must remain distinct before promotion.
- [x] 1.3 Prove one exact quantity dispute observed at threshold still creates the existing account-wide, durable, operator-only permanent block.
- [x] 1.4 Prove duplicate copies of one identity inside a comparison count once and clean comparison clears every transient streak; blank/foreign `Diff.AccountRef`, blank symbol, blank/malformed/NaN/Inf quantities remain ordinarily blocked but earn zero streak, while a valid sibling key still counts.
- [x] 1.5 Prove permanent enter failure returns with the account gate still fail-closed; then clean/different key authority-reads before withdrawing the non-durable pending row while symbol blocks remain; commit-then-error restores durable permanent, authority-read failure stays fail-closed; same key retries and becomes durable; when ordinary and permanent are both pending the permanent retry runs first; assert ordinary pending retry becomes durable and enforced.
- [x] 1.6 Run the focused tests against the frozen implementation and record the intended RED failures without changing production code.
- [x] 1.7 Prove a blank-symbol ordinary mismatch remains account-safe fail-closed through both `EntryGate` and `Tracker.EntryAllowed`, but is not persisted as an empty-symbol account-wide quantity row and cannot Restore as operator-only permanent.
- [x] 1.8 Prove `Comparer` reports 2^53-adjacent broker/local quantities as an ordinary mismatch even though they parse to one float64, while the documented `0.3`/`0.30000000000000004` round-trip artifact remains matched.
- [x] 1.9 Prove identical malformed/non-finite broker/local strings pass through the real `Comparer` as ordinary fail-closed mismatches but earn zero permanent streak evidence.
- [x] 1.10 Prove unreadable broker-only raw quantities are validated before zero/external-position classification, while a valid positive broker-only holding remains nonblocking external provenance.
- [x] 1.11 Prove a key-loss observation whose journal authority read fails retains the fail-closed gate but irreversibly breaks process-local promotion continuity; a later reappearance cannot stale-retry permanent.
- [x] 1.12 Prove audited Resolve clears a known-nondurable blank-symbol pending block without requesting release of a nonexistent journal row, while durable operator-release semantics remain unchanged.
- [x] 1.13 Prove tolerance-zero rejects decimal differences wider than one float64 ULP even when the legacy relative epsilon accepts them, while the documented 0.3 round-trip artifact remains matched.
- [x] 1.14 Prove an unrepresentable blank-symbol pending block cannot starve a valid ordinary sibling's durable write or restart projection.
- [x] 1.15 Prove the production `RawPositionsReader → Collector → Comparer → Tracker` path preserves blank/invalid holding quantity evidence and applies the same ordinary fail-closed/no-promotion contract.
- [x] 1.16 Prove one-ULP tolerance never becomes infinite at the largest finite float: exact `MaxFloat64` versus `1` is an ordinary mismatch, while exact equality still matches.
- [x] 1.17 Prove a negative finite broker-only holding is an impossible long-only projection and therefore an ordinary fail-closed mismatch, while a valid positive broker-only holding remains nonblocking external provenance.
- [x] 1.18 Prove continuity break plus authority-read failure followed by a successful `Refresh` with no durable permanent row cannot manufacture `Failures=MaxFailures` or durable/permanent state.
- [x] 1.19 Prove distinct exact integers one ULP apart (including `9007199254740992` vs `9007199254740994`) remain mismatches; only a proven short-decimal/binary-expanded round-trip pair such as `0.3`/`0.30000000000000004` may match.
- [x] 1.20 Prove an unrepresentable blank-symbol pending error cannot starve an already-earned adjusted-clean durable release of a valid sibling; the valid row disappears durably and after restart while the blank memory guard remains fail-closed.
- [x] 1.21 Prove a new ordinary mismatch observed on the comparison that breaks pending-permanent continuity is latched in memory/gate before an authority-read failure, and survives the later successful Refresh that withdraws only the stale account proposal.
- [x] 1.22 Prove holdings digest/stabilisation distinguishes blank raw quantity from exact zero in both observation orders, while two blank reads and two zero reads can each corroborate.
- [x] 1.23 Prove a later blocking comparison that refutes an adjustment credit consumes that credit before a pending-permanent authority-read error; after Refresh, a later clean read must await a new adjustment instead of releasing with stale credit.

## 2. T2 RED — missing-order and incident-chain contract

- [x] 2.1 Add canonical missing-order streak tests: identical full identity reaches threshold; changes across account/market/day/symbol/side or opaque OrderID bytes cannot share a streak; a foreign-account identity and each of the six blank components keep ordinary blocking but earn zero streak/no permanent, while a valid sibling still counts.
- [x] 2.2 Add the sanitized 2026-08-07 sequence: different symbol mismatches arrive over three cycles, earlier blocks earn adjustment credits and release, and no account-wide permanent block survives.
- [x] 2.3 Add an engine-level regression proving the changing-dispute sequence does not suppress quote/adoption after ordinary blocks clear; adoption first persists t0 and a `SEED` exit state, the intermediate view stays `not_evaluated_yet`, and a separate exit-observer judgement persists the evaluated snapshot before lines become actionable. No LIVE order adapter may be wired.
- [x] 2.4 Preserve the existing permanent-gate regression: an already durable permanent block still stops every adoption candidate before a price read.
- [x] 2.5 Run the owned focused tests against the frozen implementation and record the intended RED failures without changing production code.
- [x] 2.6 Prove two transient observations followed by Restore or Refresh do not manufacture/reconstruct a third streak; an already durable permanent row still restores as permanent.
- [x] 2.7 Prove a complete acct-7 missing-order identity inside a blank/foreign-account Diff remains an ordinary block but earns zero promotion evidence; the same identity in an acct-7 Diff still promotes normally.

## 3. T3 GREEN — minimal streak implementation

- [x] 3.1 Add an unexported comparable promotion key that uses exact finite-decimal canonicalization without float64 for quantities and requires canonical `Diff.AccountRef` equality with tracker account; reuse `LocalOrder.Identity()` normalization for orders while requiring tracker-account equality and all six canonical components non-empty; retain opaque OrderID bytes.
- [x] 3.2 Replace identity-free promotion accounting with a bounded current-dispute streak map; derive `failures` as the maximum current streak for compatibility.
- [x] 3.3 Count each identity once per comparison, drop absent identities, reset all transient streaks on non-blocking comparison, and keep invalid decimal or incomplete-order items ordinarily blocked without granting permanent-streak evidence.
- [x] 3.4 Promote through the unchanged durable account-wide operator-only path when any one exact streak reaches the existing threshold.
- [x] 3.5 Keep B8–B21 ordering intact: pre-persist gate latch, exact-cause journal writes/releases, authoritative conflict replacement and credit accounting.
- [x] 3.6 Preserve restart/refresh behavior: durable permanent rows survive; transient streaks are not reconstructed; no clean observation auto-releases permanent state.
- [x] 3.7 Bind a failed account-wide permanent write retry to its earning key's immediately consecutive blocking observation; authority-read before withdrawing on clean/key disappearance so commit-then-error restores durable state and read failure stays fail-closed; never withdraw ordinary pending or durable permanent blocks; deterministically persist simultaneous pending permanent before ordinary retries.
- [x] 3.8 Preserve exact finite quantity strings through snapshot/comparer canonicalization before promotion, and reject persistence of an unrepresentable blank-symbol ordinary quantity row while retaining its fail-closed in-memory gate.
- [x] 3.9 Reject float64-equal collisions between distinct canonical quantities in `quantitiesAgree` without removing the documented non-identical float round-trip artifact allowance.
- [x] 3.10 Validate finite decimals before equality in `quantitiesAgree`, so identical invalid/non-finite strings remain visible ordinary mismatches.
- [x] 3.11 Validate raw quantity readability before zero/external-position classification; record continuity break even when pending-permanent authority read fails; and exclude known-nondurable blank-symbol blocks from durable operator release requests while clearing them after authority succeeds.
- [x] 3.12 Replace the broad quantity relative epsilon with a one-ULP round-trip allowance, order representable ordinary writes before blank-symbol rejection, and document that the process-local key's canonical fields are copied into durable operator evidence without restoring the key map.
- [x] 3.13 Preserve raw holding quantity evidence at the production Collector boundary without changing unrelated optional-decimal blank vocabulary.
- [x] 3.14 Bound the one-ULP allowance to a finite ULP, classify negative broker-only holdings as ordinary fail-closed mismatches, and make successful `Refresh` discard a continuity-broken non-durable permanent proposal when authority has no durable permanent row.
- [x] 3.15 Restrict the ULP exception to a proven short-decimal/binary-expanded spelling pair, and persist earned representable releases before returning the known unrepresentable blank-symbol pending error.
- [x] 3.16 Latch current ordinary blocks before pending-permanent authority resolution, and use the evidence-preserving holding quantity vocabulary in the holdings digest so blank and exact zero cannot corroborate each other.
- [x] 3.17 Require the non-empty canonical Diff account, missing-order account, and tracker account to agree before a missing-order observation can earn permanent evidence.
- [x] 3.18 Apply the existing usable-credit refutation rule before the pending-permanent authority read can return an error, without spending unrelated or not-yet-usable credits.

## 4. Teammate verification and mutation ledger

- [x] 4.1 T1 and T2 rerun all owned tests GREEN against T3, then run `go test ./internal/reconcile ./internal/app/engine` with appropriate timeout.
- [x] 4.2 Run focused `-race` tests for reconcile and the a110 engine slice; run `go vet` on affected packages.
- [x] 4.3 Mutation: restore the scalar shared counter and prove the incident test RED; change quantity key to symbol-only and prove tuple test RED; drop one missing-order scope field or its non-empty validation at a time and prove canonical/incomplete-identity tests RED.
- [x] 4.4 Mutation: allow clean comparison to remove permanent state and prove existing permanent/operator tests RED; bypass ordinary block while identity is unclassifiable and prove fail-closed test RED.
- [x] 4.5 Regenerate post-edit AST/maps/risk report and make `check_analysis.py --change a110-only-the-same-dispute-becomes-permanent` pass.

## 5. Independent adversarial review

- [x] 5.1 A1 (different Terra context) reviews promotion identity, canonicalization, counter reset, duplicate handling and restart semantics; records P0/P1/P2 with concrete counterexamples.
- [x] 5.2 A2 (different Terra context) reviews durable ordering, gate scope, adoption suppression, operator-only release and exit immediacy; must try to make a false clean or false permanent state.
- [x] 5.3 T3 fixes every accepted P0/P1 with new RED-first coverage; original reviewers re-review their findings to closure.

## 6. gstack and Manager completion gate

- [x] 6.1 Run gstack code/Eng/security/QA review after adversarial closure and record decisions and any requirement edits in `review.md`.
- [x] 6.2 Run `openspec validate a110-only-the-same-dispute-becomes-permanent --strict --no-interactive`, `make sdd-sync`, `make sdd-check`, affected tests/race/vet, then `make gate CHANGE=a110-only-the-same-dispute-becomes-permanent`.
- [x] 6.3 Manager independently inspects the final diff, tasks↔design↔delta alignment, mutation evidence and frozen-base/post-edit AST changes; completion requires P1=0.
- [x] 6.4 Record the operational boundary: merge/deploy and the current block's audited release remain separate human approvals; no order or operating toggle was invoked by this change.
