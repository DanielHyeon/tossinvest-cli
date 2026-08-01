# Function Logic Map: `appendExitEventTx`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| evaluation event row | coherent saved/recomputed/effective and exact identity tuple | a042 ledger spec | SQL error aborts caller transaction |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | insert succeeds/fails | append-only event row | success or wrapped error | event roundtrip/fault test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite `ExecContext` | append history in caller transaction | no retry; caller rollback | CodeGraph + AST |

## State mutations and fallbacks

- JSON candidates are produced once by the write path and read without recomputation.

## Safety conclusion

- Safe edit boundary: additive event columns only.
- High-risk impact: yes.
