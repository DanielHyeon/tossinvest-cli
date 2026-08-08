# Branch Test Map: `NewContext`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Order path load failure | existing config/credential startup tests | baseline green | pending |
| B2 | Audit log open failure | existing audit startup tests | baseline green | pending |
| B3 | Nil clock selects system clock | existing default-clock tests | baseline green | pending |
| B4 | Gate settings audit failure | existing audit refusal tests | baseline green | pending |
| B5 | Account resolution failure | existing account startup tests | baseline green | pending |
| B6 | Journal open failure | existing durability startup tests | baseline green | pending |
| B7 | Apply-hook binding failure closes journal | existing apply-hook tests | baseline green | pending |
| B8 | Gateway construction failure closes journal | existing gateway tests | baseline green | pending |
| B9 | Gate ON + nil Guardian + production enabled enters construction | production Guardian tests | missing Guardian refused | pass |
| B10 | nil test factory selects real constructor | internal construction count test | no construction | pass |
| B11 | constructor failure closes journal and refuses | `TestProductionGuardianConstructionFailureClosesTheEngineJournal` | journal stayed unwired | pass |
| B12 | interlock failure closes journal | existing interlock suite | baseline | pass |
| B13 | unverified/gate-OFF context clears Guardian | gate-off test | baseline | pass |
| CLI | Actual assembler, one Guardian, same journal issuance/close | tagged `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver` | no Guardian injected | pass |
