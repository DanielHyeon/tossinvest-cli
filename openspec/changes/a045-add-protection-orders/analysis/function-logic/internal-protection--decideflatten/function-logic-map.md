# Function Logic Map: `decideFlatten`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | Start/deadline, authoritative decision time, exact scope+broker, cancel/sellable observations and required quantity. | Supervisor clock + broker observations | Any rollback/staleness/mismatch is IN_DOUBT with no permit. |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | validate ≤2s window; `start <= cancel <= sellable <= decisionAt <= deadline`; exact scope/broker; terminal cancel first; sufficient quantity | ALLOWED creates opaque permit with shared atomic consumption state | IN_DOUBT or decision+permit | time/identity/one-shot tables |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| scope equality + atomic permit | bind decision to one exact same-execution liquidation | no fallback; expired/replayed consume fails | CodeGraph + AST |

## State mutations and fallbacks

- Decision is otherwise pure. Successful output owns one atomic one-shot bit shared across copies.

## Safety conclusion

- Safe edit boundary: private testable core receives a clock only from the public wrapper; callers cannot supply `decisionAt` through the production API.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
