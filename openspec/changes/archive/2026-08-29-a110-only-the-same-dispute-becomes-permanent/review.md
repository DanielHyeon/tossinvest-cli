# a110 — Review ledger

## Proposal-freeze review — 2026-08-14

### Evidence convergence

| Voice | Evidence | Conclusion |
|---|---|---|
| Manager / incident | live read-only console, journal and engine logs | desired/effective adoption were ON; three different symbol mismatches incremented one scalar to 3; two ordinary rows later released while the account permanent row survived |
| Operator truth | `/position-management`, `/positions`, adoption and exit-state code | the three rows have no adoption/exit state, so `—` is truthful; adoption first opens `SEED`, and a later exit judgement creates actionable snapshot lines |
| Architecture / safety | CodeGraph impact, AST/FLM B1–B21, durable gate order | edit promotion evidence only; retain ordinary blocking, pre-persist gate latch, exact-cause release, operator-only durable permanent release and exit allowance |
| Independent Terra adversary | separate read-only context | `FIX-FIRST`, P0=0/P1=2/P2=2; findings and dispositions below |

### Findings and decisions

| Severity | Finding | Decision | Contract change |
|---|---|---|---|
| P1 | existing `canonicalDecimal` cannot report failure, accepts non-finite/malformed spellings as strings and passes through float64, so it is not proof-quality promotion identity | **ACCEPTED** | D1/D2 and delta now require a promotion-only exact finite-decimal canonicalizer; 2^53 collision, blank, malformed, NaN/Inf and mixed valid/invalid RED cases are mandatory |
| P1 | failed permanent journal enter leaves an account pending row that the generic retry loop can persist after the earning dispute disappeared | **ACCEPTED** | D3-1 binds account-pending retry to the earning key's immediately next blocking observation; clean/key disappearance withdraws only the non-durable account proposal, never ordinary pending or durable permanent rows |
| P2 | adoption itself creates a `SEED` exit state, not an immediately actionable snapshot | **ACCEPTED** | proposal D6/tasks split adoption, intermediate non-actionable view and later exit-observer evaluation |
| P2 | restart/refresh acceptance did not explicitly prove transient streak loss | **ACCEPTED** | task 2.6 adds 2-observation restart/refresh cases plus already-durable permanent restore |
| P1 (re-review) | `LocalOrder.Identity()` normalizes but does not reject blank required components, so an incomplete tuple could repeat into permanent promotion | **ACCEPTED** | D1/delta/tasks now require all six canonical order components non-empty and one RED case per missing component, including valid-sibling isolation |

### Pre-Edit Gate

- Change/task: `a110-only-the-same-dispute-becomes-permanent`, tasks 1–4.
- High-risk target: `internal/reconcile/mismatch.go`, `Tracker.Observe` promotion accounting and
  its reset/restore bookkeeping. Supporting `ReconcileDriver.blocked` is evidence only and is not an edit target.
- Caller/callee/impact evidence: CodeGraph definition/callers/callees/impact captured after `make sdd-sync`;
  impact reaches reconciliation restore tests, adoption gate and engine integration tests.
- Logic evidence: frozen AST, Function Logic Map, Branch Test Map and risk report under
  `analysis/function-logic/`; B8–B21 durable ordering is protected.
- RED-first: T1 owns promotion identity/persistence-failure tests; T2 owns missing-order and incident/adoption
  lifecycle tests; production edits wait for both intended RED reports.
- Safety: no LIVE order adapter, order preview, operating toggle, current-block release, merge, push or deploy.

### Manager freeze decision

The initial independent verdict was `FIX-FIRST`. Both P1 findings were accepted into normative design and
delta requirements. Proposal freeze becomes `ACCEPT` only after the same adversary confirms closure and
strict OpenSpec/PM validation passes. Implementation must not begin before that record is appended below.

### Closure re-review

The same independent Terra adversary re-read the revised proposal/design/delta/tasks and returned:

- original P1 exact-decimal evidence: **CLOSED**;
- original P1 stale pending-permanent retry: **CLOSED**;
- re-review P1 incomplete missing-order components: **CLOSED** after requiring all six canonical fields;
- proposal-freeze final verdict: **P0=0, P1=0 — ACCEPT**.

Manager reran strict OpenSpec validation, PM generated-tracker consistency and Function Logic Map analysis;
all passed. Task 0.7 is therefore closed and RED implementation delegation may begin.

## T1 RED adversarial review

A1 ran in a separate Terra context and returned `FIX-FIRST` with P0=1, P1=2, P2=2. Manager accepted all
actionable coverage findings: failed permanent persistence must assert the returned-state account gate;
withdrawal must assert memory, account gate and retained symbol gates; ordinary retry must assert durable
enforcement; canonical zero is pinned. The blank normalized symbol question was resolved normatively as an
unclassifiable promotion identity that remains ordinarily fail-closed but earns no permanent streak.

T1 corrected the tests in three adversarial rounds. The final additions assert:

- account gate remains permanent immediately after a failed permanent write;
- clean/different-key observation withdraws the non-durable account row from memory/gate but retains all
  ordinary symbol blocks;
- ordinary retry becomes durable and stays enforced;
- blank tracker account and blank symbol earn no promotion evidence;
- canonical zero spellings share identity;
- pending retry uses earning-key membership, not whole-diff equality, when an unrelated dispute joins.

A1 reran the focused suite and closed with **P0=0, P1=0 — ACCEPT**. The suite remains intentionally RED
against the frozen scalar implementation only at the new defect boundaries.

## T2 RED adversarial review

T2 added separate missing-order/incident and engine-chain test files. The reconcile suite proves complete
six-field missing-order identity, each blank field's no-evidence behavior, the sanitized changing-symbol
credit/release sequence, transient streak non-restoration and durable permanent restoration. The engine
suite joins that sequence to the real driver and journal.

A2 initially returned `FIX-FIRST` (P0=0, P1=1): the test stopped at adoption's `SEED` state and did not
prove the user's final line-calculation recovery. T2 then reused the real `ExitObserver` in the same incident
chain and added exact quote-batch assertions. The closed chain is now:

`ordinary release → adoption/t0 → SEED/not_evaluated_yet/— → one ExitObserver judgement → EVALUATED
canonical snapshot → shared operatorview fresh/non-dash protection line`, with zero proposal/order calls.

A2 re-reviewed and closed **P0=0, P1=0 — ACCEPT**. Its sole P2 stale comment was corrected by T2.

## T3 GREEN adversarial review

A3 ran after the first production GREEN and returned `FIX-FIRST` with P0=0/P1=3/P2=1:

- the implementation re-trimmed byte-preserving opaque `OrderID` values;
- nondeterministic map order could let a repeatedly failing ordinary pending enter preempt the required
  pending permanent retry;
- a missing order owned by another account could earn a permanent block for the tracker's account.

Manager accepted all three P1 findings into D1/D3-1, the delta and task ledger. T2 owns the opaque-ID and
foreign-account RED additions; T1 owns the simultaneous dual-pending ordering RED. T3 may fix production
only after those tests demonstrate the counterexamples. A3's P2 deterministic threshold-key selection is
recorded for traceability and must be reconsidered after the P1 fix, but it has no demonstrated unsafe
observable outcome.

The implementation and tests closed A3 at P0=0/P1=0/P2=0. The later gstack specialist pass found one
multi-specialist P1: durable permanent evidence still omitted the earning key. T1 pinned exact quantity
evidence and threshold-key handoff; T3 added deterministic quantity/missing-order evidence and corrected
the stale pooled-failure comments. Testing and security both confirmed the same journal-operability gap.

The gstack Red Team then found two more P1 boundaries: commit-then-error could be followed by an unsafe
memory-only withdrawal, and quantity promotion ignored `Diff.AccountRef`. Manager accepted both into the
normative design/delta/tasks. T1 owns RED for authoritative withdrawal and blank/foreign diff accounts;
T3 owns the minimal production correction. Final gstack verdict remains pending.

The first Codex adversarial pass found two additional production-boundary P1s that the direct Tracker tests
could not see. The real `Comparer` rounded 2^53-adjacent quantity strings through float64 before promotion,
and an ordinary blank-symbol quantity block was serialized as the same empty-symbol journal shape that
`Restore` interprets as account-wide operator-only permanent. T1 reproduced both through the real comparer
and journal/restart paths. Manager therefore amended D1/D2: exact finite canonicalization begins before Diff
construction, while an unrepresentable blank-symbol ordinary row remains fail-closed in memory but is rejected
from the ambiguous durable shape.

## gstack/Codex iterative closure

Every item below was first reproduced by a named RED in T1/T2 ownership, then changed only by T3, then
re-read by the independent Red Team. The Manager amended design/delta/tasks before each production fix.

| Boundary | Accepted correction | Closing evidence |
|---|---|---|
| float64-equal distinct exact decimals and identical invalid/non-finite spellings looked equal | exact finite canonicalization precedes equality; same-float distinct canonical strings disagree | 2^53 and invalid-string comparer REDs |
| blank/invalid raw values could be classified zero/external before validation | validate present raw values first; invalid becomes ordinary no-promotion mismatch | real Comparer raw-path REDs |
| production Collector rewrote blank raw holding to zero | raw holding quantity has a dedicated evidence-preserving canonicalizer | Collector→Comparer→Tracker RED and M19 |
| blank-symbol ordinary row could restore as permanent and block nothing in one API | reject its ambiguous journal shape; cover account-safely in both gate APIs | journal/restart and EntryAllowed REDs |
| blank pending could starve a valid sibling write | deterministic permanent→representable→blank order | sibling durability RED and M16 |
| blank pending could later starve an already-earned valid sibling release | defer the known blank error through representable additions and earned releases | durable release/history/restart RED and M24 |
| failed promotion could stale-retry after continuity loss or commit-then-timeout | break process continuity before authority read; durable authority wins, proven absence withdraws | authority outage/commit ambiguity REDs |
| a later `Refresh` could turn a continuity-broken non-durable proposal back into threshold/permanent state | successful authority with no durable row removes only that proposal and its stale scalar | Refresh recovery RED and M22 |
| pending ordinary write could preempt permanent retry | deterministic permanent-first persistence | dual-pending RED |
| durable permanent evidence omitted the earning dispute | deterministic exact earning-key detail | evidence/handoff REDs |
| missing-order account/opaque ID and incomplete fields were not proof-quality | tracker account equality, byte-preserving opaque ID, six non-empty fields | missing-order tables and per-field mutations |
| quantity `Diff.AccountRef` was not authoritative | non-empty tracker/diff account equality | blank/foreign diff-account REDs |
| broad relative epsilon/near-zero logic hid real exact differences | exact zero and explicit round-trip artifact predicate | tolerance-zero REDs |
| MaxFloat upper ULP became infinity | no generic ULP bound remains | MaxFloat RED and M20 |
| negative broker-only holding was treated as nonblocking owner exposure | ordinary fail-closed, promotion-ineligible; positive external remains unchanged | negative/positive boundary RED and M21 |
| distinct exact integers one ULP apart were treated as a round-trip artifact | accept only symmetric short fractional spelling ↔ adjacent shortest expanded spelling that rounds back at the short scale | 2^53+2 exact-integer RED, 0.3 preservation, M23 |
| authority read could fail before the current different ordinary mismatch was projected | latch current blocks and symbol gate before authority I/O; successful path skips duplicate add | authority-outage→Refresh RED and M25 |
| holdings digest aliased blank evidence with exact zero | use the dedicated holding-quantity evidence vocabulary only in the holding digest field | bidirectional stabiliser RED, equal-evidence preservation and M26 |
| complete missing-order identity could promote under a blank/foreign comparison account | require non-empty canonical tracker, Diff and order accounts to all agree | authoritative Diff-account RED/control and M27 |
| authority failure could preserve a now-refuted adjustment credit for a later clean release | refute only strictly-earlier credits for symbols disputed by the current observation before authority I/O; preserve unrelated/equal/invalid credits | outage→Refresh→clean RED and M28 |

The final artifact predicate is intentionally narrower than numerical proximity: invalid/non-finite values,
exact equality and same-float collisions are decided first; the short side must be fractional with at most
15 significant decimal digits; the expanded side must be its longer shortest float spelling, the direct
adjacent float, and round back to the short spelling at its decimal scale. Exact integers, non-shortest
spellings, non-adjacent values, MaxFloat disagreement and broad relative epsilon are rejected.

## Mutation ledger

All mutations ran only in `/tmp/tossos-a110-mutations.D1uy4T/repo`; production file hashes were compared with
the restored copy after each batch and the main worktree was never mutated by the ledger.

| Mutation | Killing test / result |
|---|---|
| M1 restore shared scalar; M2 symbol-only quantity key | changing-symbol / changed-tuple RED |
| M3a–M3f remove one missing-order validator at a time | each incomplete-field subtest RED |
| M4 let clean remove permanent; M5 bypass unclassifiable ordinary block | operator-only / fail-closed RED |
| M6 lossy float canonicalizer; M9 remove same-float collision guard; M10 accept identical invalid | exact-large / collision / invalid RED |
| M7 remove blank journal guard; M8 remove blank account-safe coverage | restart-permanent / EntryAllowed RED |
| M11 remove present-raw validation; M12 restore relative epsilon; M13 epsilon zero | unreadable external / tolerance / near-zero RED |
| M14 defer continuity break; M15 release nonexistent blank row; M16 blank-first persistence | authority outage / audited Resolve / sibling durability RED |
| M17 classify explicit local zero as external; M19 use global blank→zero Collector helper | local-zero / Collector fail-closed RED |
| M20 remove MaxFloat defense; M21 classify all broker-only as external; M22 retain stale Refresh proposal | MaxFloat / negative broker / Refresh RED |
| M23 restore generic ULP; M24 return before earned release | exact-integer / sibling release RED |
| M25 remove pre-authority ordinary latch; M26 restore holdings blank→zero digest | authority-outage gate / false-stability RED |
| M27 remove missing-order Diff-account authority check | blank/foreign comparison false-permanent RED |
| M28 remove pre-authority usable-credit refutation | outage→Refresh→clean stale-credit release RED |

M18 removed the empty-release-batch guard and **survived** because the real journal adapter treats an empty
batch as a measured no-op success. The guard remains as an explicit adapter-safety contract; the survivor is
not counted as proof of behavior unavailable at the real adapter boundary.

## Final independent verdicts

- A1/A2/A3 final: **P0=0, P1=0** after their original owners re-reviewed every accepted fix.
- gstack Red Team final current-tree verdict after F12/M28: **P0=0, P1=0, P2=0**. It verified the
  same-symbol/strictly-earlier predicate, unrelated/equal-time preservation and the absence of double effects.
- Full current-tree Codex re-review: **no actionable correctness defects identified**. Its reconcile suite and
  focused a110 engine slice passed; broader engine failures were exclusively read-only `/tmp` and listener
  restrictions, so the Manager reruns that suite in the writable verification environment below.
- No LIVE order, preview, toggle, reconcile-resolve, merge, push or deploy command was invoked.
