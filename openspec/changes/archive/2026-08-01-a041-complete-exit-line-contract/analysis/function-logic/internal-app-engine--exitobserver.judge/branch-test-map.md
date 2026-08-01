# Branch Test Map: `ExitObserver.judge`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | identity cannot be resolved | policy refusal tests | yes | yes |
| B2 | break-even cannot be priced | existing malformed-state tests | existing | yes |
| B3 | ladder receives stable observation | snapshot identity tests | yes | yes |
| B4 | ratchet receives stable observation | snapshot identity tests | yes | yes |
| B5 | successful judgement delegates without local mutation | engine integration tests | existing | yes |
