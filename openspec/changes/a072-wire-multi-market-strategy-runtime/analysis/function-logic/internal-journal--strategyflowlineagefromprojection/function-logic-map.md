# Function Logic Map: `strategyflowLineageFromProjection`

- Source: `internal/journal/strategyflow_projection.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| projection | verified sealed minor-unit evidence | strategyflow | copied exactly into payload |
| risk | canonical major-decimal Guardian intent | RiskIntent | v3 executable columns come from risk |
| schema | v2 historical or v3 current | decoded payload | select semantics explicitly |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | v2 historical schema | none | price columns use sealed minor strings | compatibility test |
| B2 | v3 current schema | none | price columns use RiskIntent major decimals | paired projection test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 | bind exact payload bytes | deterministic; no I/O | AST |
| projection getters | copy authenticated metadata | deterministic | strategyflow verifier |

## State mutations and fallbacks

- Pure value construction only; no database/network/activation side effect.
- Caller validates projection, risk, and schema first.

## Safety conclusion

- Safe edit boundary: schema-explicit durable column projection.
- High-risk impact: yes; persisted values feed first-leg verification.
