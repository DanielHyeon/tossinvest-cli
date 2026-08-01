# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed state | exact position + lifecycle generation | observer working set | stale journal refusal stops submission |
| immutable snapshot | evaluator output | exitpolicy | validation/refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | proposal/orderable combinations | builds judgement/proposal | record or no submit | exit-loop tests |
| B2 | clear/working order conditions | optional cancellation/arm suppression | error/refusal | exit-loop tests |
| B3 | journal record fails including stale generation | no broker submit | error | lifecycle race test |
| B4 | durable arm | attach/submit through gateway | result | exit e2e tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RecordExitJudgementResult` | persist and arm exact lifecycle | no retry/rebind | CodeGraph + AST |
| `submit` | broker mutation after durable authority | Guardian/gateway only | CodeGraph + AST |

## State mutations and fallbacks

- Lifecycle generation is copied from the observed state into the journal request; a stale refusal cannot reach submit.

## Safety conclusion

- Safe edit boundary: add generation attribution only; preserve record-before-submit ordering.
- High-risk impact: yes — last persistence boundary before order submission.
