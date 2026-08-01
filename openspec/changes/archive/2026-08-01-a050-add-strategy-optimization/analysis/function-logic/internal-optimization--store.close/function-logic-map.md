# Function Logic Map: `Store.Close`

- Source: `internal/optimization/store.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Store/DB pointer and sidecar paths | nil receiver, nil DB, and repeated close are harmless | Store state | return nil; close open DB once |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | nil/repeated close | no mutation beyond close state | nil | `TestCloseIsNilSafe` |
| B2 | open DB | secure existing sidecars then close DB | DB close error | lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `secureSQLiteFiles`, `sql.DB.Close` | preserve file mode and release resources | no retry | AST |

## State mutations and fallbacks

- Never touches orders, journal, lane, or gate state.

## Safety conclusion

- Safe edit boundary: private Store resource lifecycle.
- High-risk impact: no.
