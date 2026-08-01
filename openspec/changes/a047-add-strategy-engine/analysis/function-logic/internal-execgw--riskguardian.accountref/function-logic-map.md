# Function Logic Map: `RiskGuardian.AccountRef`

- Source: `internal/execgw/riskguardian.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Guardian | constructed with non-empty account | constructor | returns frozen account |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | accessor called | none | account string | adapter binding tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | immutable field accessor | no error path | AST |

## State mutations and fallbacks

- Pure accessor; no mutation or authority expansion.

## Safety conclusion

- Safe edit boundary: read-only Guardian identity.
- High-risk impact: no.
