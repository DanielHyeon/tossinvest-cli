# Branch Test Map: `ExitObserver.judgeRatchet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | policy identity mismatch | existing policy drift tests | yes | pending |
| B2 | snapshot context error | identity/decimal tests | yes | pending |
| B3 | evaluator error | existing ratchet refusal tests | yes | pending |
| B4 | unchanged snapshot | existing no-event test | yes | pending |
| B5 | changed/breach snapshot | E2E + v10 persistence test | yes | pending |
