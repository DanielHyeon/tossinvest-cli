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
