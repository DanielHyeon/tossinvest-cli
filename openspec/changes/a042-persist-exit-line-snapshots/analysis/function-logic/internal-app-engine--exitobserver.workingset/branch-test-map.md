# Branch Test Map: `ExitObserver.workingSet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | positions read error | existing loop failure test | yes | pending |
| B2 | state query/driver error | storage failure test | yes | pending |
| B3 | result indexing | account scoped test | yes | pending |
| B4 | closed/zero position | existing skip test | yes | pending |
| B5 | ineligible position | unmanaged alert test | yes | pending |
| B6 | missing state | existing open path tests | yes | pending |
| B7 | open error/completed | existing recovery tests | yes | pending |
| B8 | semantic corruption | generation quarantine test | yes | pending |
| B9 | quarantine write failure | per-position error test | yes | pending |
| B10 | valid-first ordering | blocking alert vs emergency B test | yes | pending |
| B11 | quarantine insert failure | journal error remains position-scoped | yes | yes |
| B12 | active quarantine query failure | cycle records first storage error | yes | yes |
| B13 | active generation quarantine | corrupt position is appended to refused tail | yes | yes |
| B14 | policy identity validation | existing identity conflict test | yes | yes |
| B15 | healthy state collection | emergency isolation test | yes | yes |
| B16 | refused tail append | blocking alert ordering test | yes | yes |
| B17 | valid list returned before refused list | blocking alert ordering test | yes | yes |
| B18 | unknown legacy identity | generation quarantine test | yes | yes |
| B19 | unknown-identity quarantine write failure | per-position error path | yes | yes |
| B20 | known legacy identity | pinned read-time compatibility test | yes | yes |
