# Function Logic Map: `ExitObserver.judgeLadder`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed position/state + quote | held quantity positive; runtime ID/version/digest matches executable LADDER table | journal fixed compatibility identity/registry/quote | alert and hold on refusal |
| cycle | non-nil per observation pass | `ObserveOnce` | proposal count after arm |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | registry resolution/table-fit/evaluation refuses | operator alert | nil (hold) | existing policy replacement tests |
| B2 | pending rung parses/does not parse | local pending rung or NoRung | continue | corrupt/pending tests |
| B3 | snapshot unchanged | clear refusal latch | nil | quiet rung test |
| B4 | snapshot changed | journal judgement then optional order | record result | ladder integration tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ladderFor`/`checkLadderPolicyStillFits` | immutable policy resolution and live-state defence | fail closed | CodeGraph + AST |
| exitpolicy snapshot evaluator | one authoritative transition/projection | pure/refusal | CodeGraph + AST |
| `record` | durable judgement then optional reduction | journal/broker errors returned | CodeGraph + AST |

## State mutations and fallbacks

- The evaluator receives the preserved state ID/version/digest and refuses any current-table mismatch before `record`.

## Safety conclusion

- Safe edit boundary: preserve registry/table-fit refusal and breach-first policy.
- High-risk impact: yes — exit integration; no LIVE authority change.
