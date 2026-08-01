# Branch Test Map: `appendExitEventTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | event insert success/failure | atomic evaluation fault test | yes | pending |
| B2 | evaluation payload present | snapshot persistence test | yes | yes |
| B3 | saved snapshot absent/present | first and stale recovery tests | yes | yes |
| B4 | JSON encoding failure | forged semantic output test | yes | yes |
| B5 | recomputed identity selected for event | stale recovery duplicate regression | yes | yes |
| B6 | SQL insert constraint/error | duplicate and rollback tests | yes | yes |
