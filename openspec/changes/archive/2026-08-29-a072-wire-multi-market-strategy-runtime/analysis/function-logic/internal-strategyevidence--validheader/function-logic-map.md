# Function Logic Map: `validHeader`

- Source: `internal/strategyevidence/model_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| test fixture | complete valid KR tradability header | test source of truth | callers mutate one field per refusal case |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless happy path creates the canonical fixture | none | returns a complete value | evidence model tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` / `time.Add` | build deterministic ordered source/observed/ingested times | no I/O; deterministic | AST |

## State mutations and fallbacks

- Test helper only; no production state or external source is touched.

## Safety conclusion

- Safe edit boundary: immutable test data construction.
- High-risk impact: no; coverage helper for authority validation.
