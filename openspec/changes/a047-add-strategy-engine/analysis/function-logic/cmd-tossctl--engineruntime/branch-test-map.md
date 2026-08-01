# Branch Test Map: `engineRuntime`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | invalid fill detector wiring refuses assembly | `cmd/tossctl` engine assembly/fill detector tests | baseline | baseline |
| B2 | reconcile construction error returns before start | reconcile/engine assembly tests | baseline | baseline |
| B3 | exit observer construction error returns before start | exit wiring/engine assembly tests | baseline | baseline |
| B4 | recovery construction error returns before start | recovery assembly tests plus `TestAnIncompleteRecoveryStartsNothing` | baseline | baseline |
| B5 | production assembly keeps reconcile, exit and filldetect supervised | `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver` and engine runtime tests | baseline | baseline |
| A047 | missing source/protection/provenance/scheduler keeps strategy loop absent while exit remains | dormant runtime/static tests | missing | pass |
