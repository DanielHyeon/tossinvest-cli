# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (491-578)
- AST evidence: `ast.json` — branches 6
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | assembled, verified automation context | `runEngineRun` interlock path | constructor error; no partial runtime |
| `clk` | non-nil runtime clock | caller | passed to every constructor |
| `ready` | optional post-recovery callback | boot readiness seam | nil is accepted by `recoverThenReady` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | fill detector construction fails | none | no runtime | `TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath` |
| B2 | reconcile driver construction fails | none | no runtime | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B3 | exit observer construction fails | none | no runtime | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B4 | recovery construction fails | none | no runtime | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B5 | strategy-entry construction fails | none | no runtime | current test coverage must be added before A100 edits this function |
| B6 | alert deliverer construction fails | none | no runtime | current test coverage must be added before A100 edits this function |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFillDetector` | supervised fill loop | B1 returns fail-closed | AST B1 |
| `ectx.ReconcileDriver` | supervised reconcile loop | B2 returns fail-closed | AST B2 |
| `ectx.ExitObserver` | supervised exit loop | B3 returns fail-closed | AST B3 |
| `ectx.Recovery` | recovery before loops | B4 returns fail-closed | AST B4 |
| `engine.NewRuntime` | owns loop ordering, cancellation, supervision | success only after all constructors | AST success |

## State mutations and fallbacks

- This function constructs loops; it does not start a goroutine directly.
- `Recover` is supplied to the runtime before its loop set. A100's worker must
  preserve recovery-before-cycle ordering.
- The existing worker candidates are explicitly supervised or auxiliary. A100
  must not introduce a detached protection goroutine.

## Safety conclusion

- Safe edit boundary: add the converger through `engine.NewRuntime` only after
  a verified automation gate is represented in the assembled context.
- High-risk impact: yes — a mistaken loop lifecycle can delay or block exits.
