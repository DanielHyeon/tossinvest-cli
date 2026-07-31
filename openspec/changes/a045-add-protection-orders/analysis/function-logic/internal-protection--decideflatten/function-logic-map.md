# Function Logic Map: `DecideFlatten`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Start/deadline, exact scope+broker identity, terminal cancel observation, sellable observation, required quantity. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | Existing code only checks deadline and quantity; it omits start/order/duration/scope/identity. | No mutation; returns ALLOWED or fail-closed IN_DOUBT. | Typed refusal or validated result | See branch map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | Pure time and typed equality checks. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- No mutation; returns ALLOWED or fail-closed IN_DOUBT.

## Safety conclusion

- Safe edit boundary: Require start <= cancel <= sellable <= deadline, deadline-start <=2s, and exact account/profile/market/symbol/broker on both observations.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
