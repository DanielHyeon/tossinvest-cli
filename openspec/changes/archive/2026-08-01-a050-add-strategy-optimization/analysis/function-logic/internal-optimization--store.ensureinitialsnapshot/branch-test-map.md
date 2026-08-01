# Branch Test Map: `Store.ensureInitialSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | transaction begin failure returns | DB fault review | baseline | defensive branch reviewed |
| B2 | control singleton read failure returns | store corruption coverage | pointer digest absent | PASS |
| B3 | existing pointer uses verification path | reopen tests | pointer unauthenticated | PASS |
| B4 | existing referenced snapshot corruption fails | snapshot corruption test | baseline digest hardening | PASS |
| B5 | existing control digest mismatch fails | control-pointer tamper test | rollback pointer accepted | PASS |
| B6 | empty pointer digest mismatch fails | migration/bootstrap tamper coverage | empty pointer unauthenticated | PASS |
| B7 | all registry fields considered | initial snapshot tests | baseline | PASS |
| B8 | explicit default values enter desired state only | initial descriptor tests | baseline | PASS |
| B9 | explicit effective values enter effective state only | initial descriptor tests | baseline | PASS |
| B10 | snapshot insert failure aborts | transaction fault review | baseline | defensive branch reviewed |
| B11 | control pointer CAS error aborts | transaction fault review | unauthenticated CAS | defensive branch reviewed |
| B12 | concurrent initializer losing CAS rolls back without partial state | concurrent open/bootstrap coverage | pointer CAS lacked digest | PASS |
