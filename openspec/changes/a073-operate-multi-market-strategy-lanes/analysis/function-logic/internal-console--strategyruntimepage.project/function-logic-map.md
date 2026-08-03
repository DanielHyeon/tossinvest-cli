# Function Logic Map: `strategyRuntimePage.project`

- Source: `internal/console/strategy_runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| snapshot | exactly ordered KR+US validated markets | strategy projection | caller must replace invalid input before projection |
| optional evidence/lineage times | explicit nullable values | runtime authority | rendered as `not_observed`, never zero/current fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each ordered market | append view only | none | parity/golden rendering test |
| B2 | nullable identity/time absent | presentation label only | `not_observed`/`not_configured` | dormant test |
| B3 | current vs unknown market | CSS class only | exact status retained | partial failure test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| shared field registry | stable labels/order | no failure/retry | CodeGraph + AST |
| time/string helpers | presentation only | nil remains explicit | AST |

## State mutations and fallbacks

- Appends display-only structs. It must not calculate effective state, readiness or refusal.

## Safety conclusion

- Safe edit boundary: one-way model-to-view mapping from validated shared projection.
- High-risk impact: no; no authority or side effect.
