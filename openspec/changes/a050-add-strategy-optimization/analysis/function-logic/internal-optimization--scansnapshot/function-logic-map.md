# Function Logic Map: `scanSnapshot`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selected snapshot row | required columns are canonical and digest matches | SQLite rows | corrupt snapshot error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | row scan error | none | error | DB fault path |
| B2 | JSON/invariant failure | none | corrupt error | snapshot corruption test |
| B3 | timestamp invalid | none | corrupt error | timestamp test |
| B4 | digest mismatch | none | corrupt error | digest test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseStoredTime`, `digestSnapshot` | validate immutable row | failures fail closed | AST |

## State mutations and fallbacks

- Bulk history helper avoids N+1 snapshot queries.

## Safety conclusion

- Safe edit boundary: control-store read integrity.
- High-risk impact: no.
