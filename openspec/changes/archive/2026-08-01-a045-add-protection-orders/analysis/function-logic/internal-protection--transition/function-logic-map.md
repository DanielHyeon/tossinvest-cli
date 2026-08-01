# Function Logic Map: `Transition`

- Source: `internal/protection/domain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Function inputs | A persisted-valid saga and one typed event with monotone timestamp. | Current HEAD + OpenSpec | Fail closed with typed error/decision |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1+ | invalid input/time; each state/event guard; attempt-bound registration/replace responses; broker-bound trigger/close; weaker replace; unknown mutation; output validation | copies stored input and changes state-owned fields only | typed refusal or validated result | lineage and transition tables |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Current callees | `Saga.Validate`, `invalidTransition`; pure, no broker or repository side effect. | No implicit retry; errors propagate fail-closed | CodeGraph + AST |

## State mutations and fallbacks

- Pure copy transition. Response events must prove the stored attempt/broker identity they answer.

## Safety conclusion

- Safe edit boundary: reject mismatched or irrelevant lineage fields before state mutation; retain output validation.
- High-risk impact: yes; dormant logic only, no broker mutation or WIRED binding.
