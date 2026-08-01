# Function Logic Map: `Journal.AdoptPosition`

- Source: `internal/journal/adoption.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `AdoptionRequest` | existing position, no entry decision/adoption; valid decimal t0 | `position-ledger`, schema v7 | typed invalid/not-found/already-adopted refusal |
| `positions.adoption_id` | set exactly once | v7 DDL + guarded UPDATE predicate | no row or alternate id is refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | invalid request/transaction/read failure | no durable mutation | typed/wrapped error | existing adoption validation tests |
| B4-B7 | missing, decided, or already adopted position | idempotent read only for identical derived id | fail closed otherwise | adoption tests + a044 re-adopt isolation |
| B8-B13 | insert/update/commit failures or CAS miss | transaction rollback | wrapped error | crash/migration/concurrency tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AdoptionRequest.record` | validate and derive stable adoption id | deterministic, no I/O | AST `ast.json` |
| `readAdoptionTx` | idempotent recovery | same transaction | CodeGraph + AST |
| `tx.Commit` | publish adoption and pointer atomically | FULL/WAL journal contract | journal durability tests |

## State mutations and fallbacks

- Writes one adoption row and the only guarded `positions.adoption_id` assignment.
- a044 does not edit this function or repoint `adoption_id`; re-adopt is a new lifecycle generation record.

## Safety conclusion

- Safe edit boundary: leave v7 adoption/t0 immutable; add a separate generation-scoped repository.
- High-risk impact: yes — adoption establishes protection t0; regression tests must prove no broker/order effect.
