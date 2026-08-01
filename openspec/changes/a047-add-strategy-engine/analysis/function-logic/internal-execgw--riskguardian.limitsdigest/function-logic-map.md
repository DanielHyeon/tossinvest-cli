# Function Logic Map: `RiskGuardian.LimitsDigest`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Guardian limits JSON | constructor-derived canonical snapshot | Guardian | SHA-256 returned |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | accessor called | local hash only | lowercase hex digest | adapter binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| SHA-256 | bind canonical limits JSON | deterministic/no error | AST |

## State mutations and fallbacks

- Pure digest; does not expose mutable limits.

## Safety conclusion

- Safe edit boundary: read-only activation binding.
- High-risk impact: no direct mutation.
