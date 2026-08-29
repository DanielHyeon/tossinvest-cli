# Function Logic Map: `insertFirstLegBindingTx`

- Source: `internal/journal/strategy_first_leg_atomic.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| prepared first leg | fully validated q_final and optional weekly reservation | preparation | transaction rollback on insert error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | immutable first-leg insert fails | transaction-local writes | wrapped error | late failpoint suite |
| B2 | weekly companion insert fails | transaction-local writes | wrapped error | paired weekly mismatch/rollback |
| success | required rows insert | immutable companions only | nil | paired weekly success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQL inserts | bind general first leg and optional weekly key in same transaction | no internal retry | CodeGraph + AST |

## State mutations and fallbacks

- Writes no lease or broker record; all inserts roll back with the owning BEGIN IMMEDIATE transaction.

## Safety conclusion

- Safe edit boundary: add the weekly companion before commit without independent transaction or fallback.
- High-risk impact: yes/no
