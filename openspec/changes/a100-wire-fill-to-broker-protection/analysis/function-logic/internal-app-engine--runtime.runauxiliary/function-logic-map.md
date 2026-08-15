# Function Logic Map: `Runtime.runAuxiliary`

- Source: `internal/app/engine/auxiliary.go` (87-114)
- AST evidence: `ast.json` — AST branches 2.

## Inputs and invariants

An auxiliary is started by `Runtime.Run` in its shared wait group but never sends to the supervised `stops` channel. It receives `loopCtx`, not the parent context. Thus a supervised-loop failure cancels the auxiliary and is a graceful stop, while an independent auxiliary failure is observable without stopping engine loops.

## Branches and early returns

| Branch | Result | Existing / planned test |
|---|---|---|
| B1 | context-cancelled auxiliary exits as graceful, no gate latch | `TestCancellingTheRuntimeDoesNotLockTheEntryGate` |
| B2 | absent `OnStop` logs only and returns | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |

## Calls and live bindings

Calls `runAuxiliaryBody`, `gracefulStop`, `r.log`, and optional `aux.OnStop`. Its non-graceful event is `EventAlertUndelivered`, deliberately not an engine-loop stop event. The worker must not use auxiliary semantics if it is required to participate in the engine's stop event or needs a circular gate-latch callback.

## State mutations and fallbacks

No direct journal/order mutation. The optional callback may latch external state; that callback occurs only after an independently failed auxiliary. No restart occurs.

## Safety conclusion

This is the A0 circularity seam: wiring a protection worker here with `OnStop` that changes entry authority can cause the worker's own failure path to become a gate transition. Use a non-circular supervised composition or prove the callback cannot feed the worker's start condition.
