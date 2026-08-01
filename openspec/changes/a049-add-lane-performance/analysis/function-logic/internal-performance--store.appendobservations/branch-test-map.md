# Branch Test Map: `Store.AppendObservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | begin failure returns without write | context/error coverage | yes | yes |
| B2 | absent ID appends | `TestObservationCompareAndAppendReplayAndDivergence` | yes | yes |
| B3 | identical ID replay skips | replay/restart/concurrent tests | yes | yes |
| B4 | divergent ID fails closed | replay/concurrent tests | yes | yes |
| B5 | SIGKILL mid-transaction is all-or-none | `TestPerformanceMigrationAndAppendSIGKILLPhasesAreAllOrNone` | yes | yes |
