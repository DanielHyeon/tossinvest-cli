# Branch Test Map: `Runtime.runAuxiliary`

Source: `internal/app/engine/auxiliary.go` (87-114).

| Branch | Scenario | Test |
|---|---|---|
| B1 | runtime cancellation is graceful and does not latch entry | `TestCancellingTheRuntimeDoesNotLockTheEntryGate` |
| B2 | dead auxiliary with no callback is logged but does not stop loops | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |

Planned A100 RED: worker failure cannot create an entry-gate state that prevents the same recovery topology from starting; loop cancellation still drains it before runtime return.
