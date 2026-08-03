# Function Logic Map: `Console.strategyRuntimeSummary`

- Source: `internal/console/settings_tabs.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| reader | nil or read-only paired projection | Console.Options | nil/error/invalid described as unavailable |
| markets | exact KR+US | shared projection | summary never substitutes one market for another |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | reader nil | none | paired dormant summary | settings summary test |
| B2 | reader error/invalid | none | typed read failure summary | settings summary test |
| B3 | valid pair | none | KR and US desired/effective/refusal summary | settings summary test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyRuntimeReader.Read` | obtain same model as runtime page | single read, no retry/mutation | CodeGraph + AST |
| shared validator | prevent partial trust | invalid whole payload becomes typed failure | AST |

## State mutations and fallbacks

- Returns text only; no setting or runtime mutation.

## Safety conclusion

- Safe edit boundary: summarize both market records without recomputing effective state.
- High-risk impact: no; read-only navigation text.
