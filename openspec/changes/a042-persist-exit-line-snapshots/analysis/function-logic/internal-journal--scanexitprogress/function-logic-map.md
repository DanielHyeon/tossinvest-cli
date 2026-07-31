# Function Logic Map: `scanExitProgress`

- Source: `internal/journal/exit_state.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| tx + position id | existing exit state, nullable v10 snapshot tuple | journal schema v10 | typed not-found/storage error; semantic evidence preserved |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | no row | none | `ErrExitStateNotFound` | existing tests |
| B2 | SQL scan error | none | wrapped storage error | fault test |
| B3 | nullable rung/effective snapshot | populate typed current state | success/semantic validation later | v10 roundtrip/legacy tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SQLite `QueryRowContext` | read inside judgement transaction | no retry; caller rolls back | CodeGraph + AST |

## State mutations and fallbacks

- Nullable columns use `sql.Null*`; no coalesce or registry backfill.

## Safety conclusion

- Safe edit boundary: transaction-local read only.
- High-risk impact: yes.
