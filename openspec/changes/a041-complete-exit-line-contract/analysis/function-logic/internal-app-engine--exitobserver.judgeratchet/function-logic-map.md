# Function Logic Map: `ExitObserver.judgeRatchet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position/state + quote/break-even | held quantity positive; runtime identity equals active ratchet semantics | journal/quote/cost model + pinned legacy identity | alert and hold on refusal |
| cycle | non-nil per observation pass | `ObserveOnce` | proposal count only after durable arm |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | active config identity differs from state | latched operator alert | nil (hold) | policy reinterpretation tests |
| B2 | snapshot context/evaluator refuses | latched operator alert | nil (hold) | refusal tests |
| B3 | immutable snapshot has no state change | clear refusal latch only | nil | quiet observation tests |
| B4 | snapshot changed | journal judgement then optional order | record result | exitloop integration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RatchetPolicyIdentity`/snapshot evaluator | exact identity gate then one authoritative decision/projection | pure/refusal | CodeGraph + AST |
| `record` | durable judgement then optional reduction | journal/broker errors returned | CodeGraph + AST |

## State mutations and fallbacks

- The fixed runtime identity is compared before evaluation; the resulting snapshot is passed whole to `record`.

## Safety conclusion

- Safe edit boundary: retain alert/hold semantics and append-only quiet-cycle suppression.
- High-risk impact: yes — exit integration; no LIVE authority change.
