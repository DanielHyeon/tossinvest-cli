# Function Logic Map: `ExitObserver.judgeLadder`

- Source: `internal/app/engine/exitloop.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| managed ladder state + quote + observation | exact immutable ladder and valid active rung/pending state | a041 evaluator | refuse/alert without order on mismatch |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B7 | policy resolve, table fit, context, pending rung, evaluate, no-change, record | clear refusal or persist exact snapshot | nil/refusal/record error | ladder drift/snapshot/E2E tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `EvaluateLadderSnapshot` | single calculation shared by persistence/execution | pure; errors fail closed | CodeGraph + AST |
| `record` | atomic journal then optional submit | commit-before-submit | CodeGraph + AST |

## State mutations and fallbacks

- Passes original quote observation metadata with the exact evaluated snapshot.
- The exact `LadderSnapshotInput` and previous watermark/protection/rung are persisted as recovery evidence; the rung table is defensively copied.

## Safety conclusion

- Safe edit boundary: no policy arithmetic change.
- High-risk impact: yes.
