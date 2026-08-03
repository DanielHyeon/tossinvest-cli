# Branch Test Map: `runVerifyAbort`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `--list` sends no request | existing verify abort list coverage | existing | existing |
| B2 | unreadable list record refuses before live access | existing verify abort record coverage | existing | existing |
| B3 | nil command context falls back safely | command wiring tests | existing | existing |
| B4 | active verification refuses abort without erasing its marker | `TestA061ActiveVerificationExclusionRefusesAbortWithoutErasingItsMarker` | yes | yes |
| B5 | profile intent marker path fails | profile path tests | existing | existing |
| B6 | occupied metadata lease blocks broker while marker stays fresh | `TestA061RunVerifyAbortPublishesItsMarkerBeforeWaitingForTheProfileBudget` | yes | yes |
| B7 | record is reloaded only after exclusive admission | `TestA061AbortReloadsOutstandingTargetsAfterExclusiveAdmission` | yes | yes |
| B8 | refreshed record with no target returns without broker | existing empty-record abort coverage | existing | existing |
| B9 | broker failure follows refreshed target display | refreshed-target test | yes | yes |
| B10 | recorder open failure releases exclusions | existing record permission tests | existing | existing |
| B11 | invalid runner setup releases exclusions | verifylive option tests | existing | existing |
| B12 | declined abort reason is surfaced | existing abort approval tests | existing | existing |
| B13 | approved cleanup with no remaining target reports completion | verifylive abort tests | existing | existing |
| B14 | context termination preserves evidence | existing abort interrupt tests | existing | existing |
