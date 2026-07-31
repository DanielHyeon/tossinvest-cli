# Branch Test Map: `ExitObserver.record`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero projected partial persists promotion, no proposal/reservation/broker | a041 zero-order integration test | no | yes |
| B2 | full exit clears symbol then arms/submits | existing liquidation integration tests | existing | yes |
| B3 | uncleared symbol advances state but withholds proposal | existing bounded-delay tests | existing | yes |
| B4 | concurrent duplicate consumer gets `ErrProposalPending`, one submission | a041 race test | no | yes |
| B5 | journal failure prevents submission | existing crash-ordering tests | existing | yes |
| B6 | clear-symbol failure returned | existing cancellation failure tests | existing | yes |
| B7 | clear-symbol success clears delay | existing liquidation tests | existing | yes |
| B8 | orderable mints intent/proposal | existing proposal tests | existing | yes |
| B9 | journal pending race is benign | a041 duplicate-consumer race | no | yes |
| B10 | judgement has no proposal | a041 zero-order test | no | yes |
| B11 | proposal armed increments cycle and submits | existing exit submission tests | existing | yes |
