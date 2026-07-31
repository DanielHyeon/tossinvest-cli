# Function Logic Map: `ExitObserver.judgeRatchet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position/state + price/break-even | held quantity positive; coherent RATCHET state | journal/quote/cost model | alert and hold on refusal |
| cycle | non-nil per observation pass | `ObserveOnce` | proposal count only after durable arm |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | evaluator/snapshot refuses | latched operator alert | nil (hold) | refusal tests |
| B2 | immutable snapshot has no state change | clear refusal latch only | nil | quiet observation tests |
| B3 | snapshot changed | journal judgement then optional order | record result | exitloop integration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| exitpolicy snapshot evaluator | one authoritative decision/projection | pure/refusal | CodeGraph + AST |
| `record` | durable judgement then optional reduction | journal/broker errors returned | CodeGraph + AST |

## State mutations and fallbacks

- Current function re-materialises `ExitJudgement` and sends a separate proposal. Edit will pass one snapshot whose fields drive both.

## Safety conclusion

- Safe edit boundary: retain alert/hold semantics and append-only quiet-cycle suppression.
- High-risk impact: yes — exit integration; no LIVE authority change.
