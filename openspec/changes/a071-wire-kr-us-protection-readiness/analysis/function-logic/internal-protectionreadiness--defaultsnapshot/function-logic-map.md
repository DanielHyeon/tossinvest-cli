# Function Logic Map: `DefaultSnapshot`

- Source: `internal/protectionreadiness/readiness.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| paired release | exact KR and US | compiled release | both UNWIRED |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | branchless construction | local seals only | paired snapshot | default assembly test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| market/global seal helpers | immutable defaults | no I/O | AST |

## State mutations and fallbacks

- Constructs both market refusals; no provider, broker, toggle or approval mutation.

## Safety conclusion

- Safe edit boundary: add independent market seals.
- High-risk impact: no; defaults remain UNWIRED.
