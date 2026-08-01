# Function Logic Map: `Journal.OpenExitStateResults`

- Source: `internal/journal/exit_snapshot.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | trimmed account scope | engine context | query error, never widen scope |
| lifecycle row | current MANAGED generation only | v12 ledger | released/old generations omitted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | query fails | none | wrapped error | persistence tests |
| B2 | scan fails/corrupt | none | error/result corruption | snapshot tests |
| B3 | iterator fails | none | wrapped error | snapshot tests |
| B4 | valid row | append result | results | engine lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `db.QueryContext` | select authoritative observer set | fail closed, no retry | CodeGraph + AST |
| `scanExitStateResult` | hydrate snapshot/corruption | error propagated | CodeGraph + AST |

## State mutations and fallbacks

- Read-only. Lifecycle join must not convert absence into implicit management after a release.

## Safety conclusion

- Safe edit boundary: mirror `OpenExitStates` lifecycle predicates exactly.
- High-risk impact: yes — production observer uses this richer query.
