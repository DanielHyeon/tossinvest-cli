# Branch Test Map: `Runtime.Run`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | recovery failure starts no loop | `TestAnIncompleteRecoveryStartsNothing` | baseline | baseline |
| B2 | recovery completes before every loop | `TestRecoveryRunsBeforeAnyLoopStarts` | baseline | baseline |
| B3 | caller cancellation drains every loop without critical | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical` | baseline | baseline |
| B4 | nil/error return or shutdown race stops peers and alerts | `TestALoopReturningOnItsOwnStopsEverythingAndIsCritical`, `TestALoopThatSimplyReturnsIsAlsoAFailure`, `TestALoopFailingDuringAShutdownIsStillReported` | baseline | baseline |
| A047 | lane OFF produces no entry and adds no entry loop while existing exit/reconcile remains unchanged | dormant wiring/static tests | missing | pass |
