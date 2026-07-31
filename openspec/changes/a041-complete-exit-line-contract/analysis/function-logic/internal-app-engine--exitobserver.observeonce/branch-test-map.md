# Branch Test Map: `ExitObserver.ObserveOnce`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | fill detection pressure defers | existing SLO tests | existing | yes |
| B2 | working set fails | existing journal error tests | existing | yes |
| B3 | empty account resets outage | existing empty-account tests | existing | yes |
| B4 | observation fails | existing outage tests | existing | yes |
| B5 | quote omits a held symbol | existing partial answer tests | existing | yes |
| B6 | concurrent consumers judge one fetched quote | `TestExitObserverConcurrentSameQuoteArmsAndMutatesOnce` | yes | yes |
| B7 | cycle keeps first per-position error while continuing | existing multi-position tests | existing | yes |
