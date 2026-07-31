# Function Logic Map: `ExitObserver.judge`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed state + observed quote | eligible state with immutable runtime identity | working set + quote source | refusal alert and hold |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | break-even cannot be computed | refusal latch | hold | existing malformed-position tests |
| B2 | ladder state | none directly | delegate | ladder tests |
| B3 | ratchet/default state | none directly | delegate | ratchet tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `breakEven` | cost-inclusive protection candidate | error refuses judgement | CodeGraph + AST |
| `judgeLadder`/`judgeRatchet` | authoritative evaluator path | child handles refusal | CodeGraph + AST |

## State mutations and fallbacks

- Threads the same quote metadata and fallback cycle observation to the selected evaluator.

## Safety conclusion

- Safe edit boundary: no new policy branch; only typed observation propagation.
- High-risk impact: yes — evaluator dispatch.
