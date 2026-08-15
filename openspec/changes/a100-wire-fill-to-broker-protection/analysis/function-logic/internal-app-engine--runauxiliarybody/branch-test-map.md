# Branch Test Map: `runAuxiliaryBody`

Source: `internal/app/engine/auxiliary.go` (123-130).

| Branch | Scenario | Test |
|---|---|---|
| B1 | panic is recovered into an auxiliary stop | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` |

Planned A100 RED: a worker panic remains observable and does not silently restart, stop unrelated loops, or leave a circular entry latch.
