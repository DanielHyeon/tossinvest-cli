# Branch Test Map: `reconcileStateID`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | identical time/symbol across KR/US yields distinct IDs | market coexistence test | pending | pending |
| B2 | length-prefix hash loop includes every component | market coexistence and re-entry tests | yes | yes |
