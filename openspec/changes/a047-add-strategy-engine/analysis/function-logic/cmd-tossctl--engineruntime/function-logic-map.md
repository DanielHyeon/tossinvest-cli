# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
| --- | --- | --- | --- |
| `ectx` | Fully assembled production engine context | `assembleEngine` startup interlock | A missing downstream dependency returns an assembly error before runtime start. |
| `clk` | One clock shared by detector, reconciliation, exit and recovery | command wiring | Construction fails or the component's documented default applies; a047 must not introduce a second time authority. |
| `logger` | Optional structured logger | command wiring | Runtime remains functional without changing safety state. |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
| --- | --- | --- | --- | --- |
| B1 | fill detector construction fails | none | error; no other loop is built | engine assembly refusal tests |
| B2 | reconcile driver construction fails | none | error; runtime is absent | engine assembly tests |
| B3 | exit observer construction fails | none | error; runtime is absent | exit wiring tests |
| B4 | recovery construction fails | none | error; no loop starts | recovery assembly tests |
| B5 | all components valid | constructs exactly reconcile, exit and filldetect loop descriptors | `engine.NewRuntime(...)` result | `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver` and runtime loop-name tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
| --- | --- | --- | --- |
| `engineFillDetector` | official-read fill observation | assembly error is terminal; polling retry belongs to detector | CodeGraph + AST |
| `Context.ReconcileDriver` | account/journal convergence | construction fail closes startup | CodeGraph + AST |
| `Context.ExitObserver` | preserve existing position exits | construction fail closes startup; no strategy dependency | CodeGraph + AST |
| `Context.Recovery` | resolve durable attempts before loops | failure later stops start in `Runtime.Run` | CodeGraph + AST |
| `engine.NewRuntime` | validate and supervise loop set | rejects invalid/duplicate/unsupervisable loop wiring | CodeGraph + AST |

## State mutations and fallbacks

- Only allocates components; it performs no order mutation.
- The existing three-loop set is the baseline. a047 cannot remove, condition, or
  delay the exit/reconcile/filldetect loops when strategy entry is OFF.
- No fallback starts a partial runtime after an assembly error.

## Safety conclusion

- Safe edit boundary: add an entry loop only after a045/a046/a048/source
  readiness is structurally available and independently validated; otherwise
  leave this function unchanged and expose a dormant component.
- High-risk impact: yes.
