# Function Logic Map: `ExitObserver.judgeRatchet`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed state + quote + cycle observation | stored policy identity equals active immutable ratchet; decimal state valid | a041 evaluator | refuse/alert without order on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | identity resolve/match, context creation, evaluate, no-change, record | clear refusal or persist exact snapshot | nil/refusal/record error | ratchet identity/snapshot tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EvaluateRatchetSnapshot` | single calculation shared by persistence/execution | pure; errors fail closed | CodeGraph + AST |
| `record` | atomically persist then optionally submit | journal commit precedes broker | CodeGraph + AST |

## State mutations and fallbacks

- Only the `record` call gains observation metadata; evaluator and emergency branch are unchanged.
- The exact `RatchetSnapshotInput` is retained and copied into recovery evidence together with the previous watermark/protection/level; recovery therefore reruns the same evaluator input.

## Safety conclusion

- Safe edit boundary: pass already-captured quote timestamp/source to journal.
- High-risk impact: yes.
