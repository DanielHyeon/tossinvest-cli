# Function Logic Map: `Runtime.Run`

- Source: `internal/app/engine/runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| caller context | live until deliberate shutdown | `runEngineRun` | cancellation drains every loop |
| `opts.Recover` | nil in isolated tests or complete recovery function | runtime assembly | recovery error starts zero loops |
| `opts.Loops` | prevalidated non-empty unique loop set | `NewRuntime` | invalid wiring is rejected before `Run` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | recovery exists and fails | recovery only; no goroutine starts | recovery error | `TestAnIncompleteRecoveryStartsNothing` |
| B2 | recovery absent or succeeds | starts all loops and health supervisor | waits for first loop stop | `TestRecoveryRunsBeforeAnyLoopStarts` |
| B3 | first stop is caused by caller cancellation | cancels/drains peers, heartbeat log | nil | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical` |
| B4 | loop returns nil/other error, including shutdown race | cancels/drains peers, critical alert and log | wrapped `ErrLoopFailed` | loop-return and shutdown-race tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `Recover` | durable restart convergence | one failure is terminal; never retried here | CodeGraph + AST |
| each `SupervisedLoop.Run` | operate runtime component | loop owns cycle retry; supervisor never restarts an unexplained return | CodeGraph + AST |
| `superviseHealth` | escalate sustained degradation | polls until runtime cancellation | CodeGraph + AST |
| `drain`, `gracefulStop`, `alert` | enforce no partial survival | waits for all goroutines before return | CodeGraph + AST |

## State mutations and fallbacks

- Starts goroutines only after recovery succeeds.
- The first stopped loop cancels all other loops and the health supervisor.
- No partial-runtime or automatic-restart fallback exists.

## Safety conclusion

- Safe edit boundary: a strategy loop, if eventually supplied, must follow the
  same cancellation and failure rules; OFF behavior belongs inside the entry
  component and must not short-circuit other loops.
- High-risk impact: yes.
